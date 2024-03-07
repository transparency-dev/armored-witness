// Copyright 2023 The Armored Witness authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Verify is a tool for inspecting the installed firmware & configuration
// ArmoredWitness devices, and checking that all installed firmware images
// are preset in firmware transparency log(s).
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"time"

	"k8s.io/klog/v2"

	"github.com/transparency-dev/armored-witness-boot/config"
	"github.com/transparency-dev/armored-witness-common/release/firmware"
	"github.com/transparency-dev/armored-witness-common/release/firmware/update"
	"github.com/transparency-dev/armored-witness/internal/device"
	"github.com/transparency-dev/armored-witness/internal/fetcher"
	"github.com/transparency-dev/formats/log"
	"github.com/transparency-dev/merkle/proof"
	"github.com/transparency-dev/merkle/rfc6962"
	"github.com/transparency-dev/serverless-log/client"
	"golang.org/x/mod/sumdb/note"
)

const (
	// Block size in bytes of the MMC device on the armored witness.
	mmcBlockSize = 512

	// bootloaderConfigBlock defines the location of the bootloader config GOB on MMC.
	// In constrast to the other firmware binaries below where each firmware is preceeded
	// by its config GOB, the bootloader config is stored separatly due to the hard requirement
	// for the binary location imposed by the i.MX ROM bootloader.
	bootloaderConfigBlock = 0x4FB0
	// osBlock defines the location of the first block of the TrustedOS on MMC.
	osBlock = 0x5000
	// appletBlock defines the location of the first block of the TrustedApplet on MMC.
	appletBlock = 0x200000
)

var (
	firmwareLogURL      = flag.String("firmware_log_url", "", "URL of the firmware transparency log to scan for firmware artefacts.")
	firmwareLogOrigin   = flag.String("firmware_log_origin", "", "Origin string for the firmware transparency log.")
	firmwareLogVerifier = flag.String("firmware_log_verifier", "", "Checkpoint verifier key for the firmware transparency log.")
	binariesURL         = flag.String("binaries_url", "", "Base URL for fetching firmware artefacts referenced by FT log.")

	appletVerifier   = flag.String("applet_verifier", "", "Verifier key for the applet manifest.")
	bootVerifier     = flag.String("boot_verifier", "", "Verifier key for the boot manifest.")
	osVerifier1      = flag.String("os_verifier_1", "", "Verifier key 1 for the OS manifest.")
	osVerifier2      = flag.String("os_verifier_2", "", "Verifier key 2 for the OS manifest.")
	recoveryVerifier = flag.String("recovery_verifier", "", "Verifier key for the recovery manifest.")

	blockDeviceGlob = flag.String("blockdevs", "/dev/sd*", "Glob for plausible block devices where the armored witness could appear.")

	runAnyway = flag.Bool("run_anyway", false, "Let the user override bailing on any potential problems we've detected.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	v := verifierFromFlags()

	ctx := context.Background()

	if u, err := user.Current(); err != nil {
		klog.Exitf("Failed to determine who I'm running as: %v", err)
	} else if u.Uid != "0" {
		klog.Warningf("⚠️ This tool probably needs to be run as root (e.g. via sudo), it's running as %q (UID %q); re-run with the --run_anyway flag if you know better.", u.Username, u.Uid)
		if !*runAnyway {
			klog.Exit("Bailing.")
		}
	}

	if err := v.waitAndVerify(ctx); err != nil {
		klog.Exitf("❌ Failed to verify device: %v", err)
	}
	klog.Info("✅ Device verified OK!")
	klog.Info("----------------------------------------------------------------------------------------------")
	klog.Info("🙏 Operator, please ensure boot switch is set to MMC, and then reboot device 🙏")
	klog.Info("----------------------------------------------------------------------------------------------")
}

// firmwares respresents the collection of firmware and related artefacts found
// on the device.
type firmwares struct {
	// Bootloader holds the regular bootloader firmware bundle.
	Bootloader firmware.Bundle
	// TrustedOS holds the trusted OS firmware bundle.
	TrustedOS firmware.Bundle
	// TrustedApplet holds the witness applet firmware bundle.
	TrustedApplet firmware.Bundle
}

// verifier is a struct which knows how to verify firmware transparency inclusion for
// firmware on an armored witness device.
type verifier struct {
	logOrigin string
	logV      note.Verifier

	bootV     note.Verifier
	appletV   note.Verifier
	osV1      note.Verifier
	osV2      note.Verifier
	recoveryV note.Verifier

	logBaseURL *url.URL
	binBaseURL *url.URL

	recovery firmware.Bundle
}

// fetchRecoveryFirmware returns a recovery image suitable for use on the armored witness,
// and which has been verified to be present in the firmware transparency log.
//
// TODO: this will need updating to fetch a specific version of the image which has
// been signed for the attached device.
func (v *verifier) fetchRecoveryFirmware(ctx context.Context) error {
	logFetcher := fetcher.New(v.logBaseURL)
	binFetcher := fetcher.BinaryFetcher(logFetcher)
	updateFetcher, err := update.NewFetcher(ctx,
		update.FetcherOpts{
			LogFetcher:       logFetcher,
			LogOrigin:        v.logOrigin,
			LogVerifier:      v.logV,
			BinaryFetcher:    binFetcher,
			AppletVerifier:   v.appletV,
			BootVerifier:     v.bootV,
			OSVerifiers:      [2]note.Verifier{v.osV1, v.osV2},
			RecoveryVerifier: v.recoveryV,
		})
	if err != nil {
		return fmt.Errorf("NewFetcher: %v", err)
	}

	if err := updateFetcher.Scan(ctx); err != nil {
		return fmt.Errorf("Scan: %v", err)
	}

	r, err := updateFetcher.GetRecovery(ctx)
	if err != nil {
		return fmt.Errorf("GetRecovery: %v", err)
	}

	bv := firmware.BundleVerifier{
		LogOrigin:         v.logOrigin,
		LogVerifer:        v.logV,
		ManifestVerifiers: []note.Verifier{v.recoveryV},
	}

	if _, err := bv.Verify(r); err != nil {
		return err
	}

	v.recovery = r
	return nil
}

// waitAndVerify attempts to boot a connected armored witness device into recovery mode,
// directly extracts the bootloader, trusted OS, and trusted applet firmware from the MMC,
// before finally verifying that the images and manifests are self-consistent, the manifest
// is signed by the correct key(s), the manifests are present in the firmware transparency log,
// and that the bundled log checkpoint is consistent with the current view of the log from the
// workstation running this command.
func (v *verifier) waitAndVerify(ctx context.Context) error {
	if err := v.fetchRecoveryFirmware(ctx); err != nil {
		klog.Exitf("Failed to fetch device recovery image: %v", err)
	}
	klog.Info("Successfully fetched and verified recovery image")
	klog.Info("----------------------------------------------------------------------------------------------")
	klog.Info("🙏 Operator, please ensure boot switch is set to USB, and then connect unprovisioned device 🙏")
	klog.Info("----------------------------------------------------------------------------------------------")

	recoveryHAB := append(v.recovery.Firmware, v.recovery.HABSignature...)
	klog.Infof("Recovery firmware is %d bytes + %d bytes HAB signature", len(v.recovery.Firmware), len(v.recovery.HABSignature))
	// The device will initially be in HID mode (showing as "RecoveryMode" in the output to lsusb).
	// So we'll detect it as such:
	target, bDev, err := device.BootIntoRecovery(ctx, recoveryHAB, *blockDeviceGlob)
	if err != nil {
		return err
	}
	klog.Infof("✅ Detected device %q", target.DeviceInfo.Path)
	klog.Infof("✅ Detected blockdevice %v", bDev)

	var fw *firmwares
	// There appears to be a race on Linux between the device file appearing and being able to open and use it.
	// Give it a couple of tries just in case:
	for i := 0; i < 2; i++ {
		fw, err = extractFirmware(bDev)
		if err != nil {
			klog.Warningf("Failed to extract firmware: %v", err)
			time.Sleep(time.Second)
			klog.Info("Retrying...")
			continue
		}
		break
	}
	if err != nil {
		return fmt.Errorf("failed to extract firmware after multiple attempts: %v", err)

	}

	if err := v.verifyFirmwares(ctx, *fw); err != nil {
		return err
	}

	return nil
}

// verifyFirmwares performs the firmware transparency verification of the firmware bundles
func (v *verifier) verifyFirmwares(ctx context.Context, fw firmwares) error {
	logFetcher := fetcher.New(v.logBaseURL)
	lst, err := client.NewLogStateTracker(ctx, logFetcher, rfc6962.DefaultHasher, nil, v.logV, v.logOrigin, client.UnilateralConsensus(logFetcher))
	if err != nil {
		return fmt.Errorf("failed to create LogStateTracker: %v", err)
	}

	errs := []error{}

	for _, p := range []struct {
		name       string
		bundle     firmware.Bundle
		manifestVs []note.Verifier
	}{
		{name: "Bootloader", bundle: fw.Bootloader, manifestVs: []note.Verifier{v.bootV}},
		{name: "TrustedOS", bundle: fw.TrustedOS, manifestVs: []note.Verifier{v.osV1, v.osV2}},
		{name: "TrustedApplet", bundle: fw.TrustedApplet, manifestVs: []note.Verifier{v.appletV}},
	} {
		// First verify that the stored proof bundle is self-consistent:
		bv := firmware.BundleVerifier{
			LogOrigin:         v.logOrigin,
			LogVerifer:        v.logV,
			ManifestVerifiers: p.manifestVs,
		}
		extractedFWHash := sha256.Sum256(p.bundle.Firmware)
		klog.V(1).Infof("%s extracted firmware has base64 hash: %s", p.name, base64.StdEncoding.EncodeToString(extractedFWHash[:]))
		klog.V(1).Infof("%s Manifest:\n%s", p.name, p.bundle.Manifest)
		if _, err := bv.Verify(p.bundle); err != nil {
			klog.Infof("  ❌ %s: %v", p.name, err)
			errs = append(errs, fmt.Errorf("failed to verify %s: %v", p.name, err))
			continue
		}
		klog.Infof("  ✅ %s: proof bundle is self-consistent ", p.name)

		// Now verify that the checkpoint used in the proofbundle is consitent with our
		// view of the log:
		fwCP, _, _, err := log.ParseCheckpoint(p.bundle.Checkpoint, v.logOrigin, v.logV)
		if err != nil {
			return fmt.Errorf("failed to parse checkpoint from %s: %v", p.name, err)
		}
		if fwCP.Size > lst.LatestConsistent.Size {
			if _, _, _, err := lst.Update(ctx); err != nil {
				return fmt.Errorf("failed to update LogStateTracker: %v", err)
			}
		}

		cp, err := lst.ProofBuilder.ConsistencyProof(ctx, fwCP.Size, lst.LatestConsistent.Size)
		if err != nil {
			return fmt.Errorf("failed to build consistency proof for %s checkpoint: %v", p.name, err)
		}
		if err := proof.VerifyConsistency(rfc6962.DefaultHasher, fwCP.Size, lst.LatestConsistent.Size, cp, fwCP.Hash, lst.LatestConsistent.Hash); err != nil {
			klog.Infof("%s proof bundle checkpoint:\n%s", p.name, p.bundle.Checkpoint)
			klog.Infof("%s my checkpoint:\n%s", p.name, lst.LatestConsistentRaw)
			return fmt.Errorf("invalid consistency proof for %s checkpoint: %v", p.name, err)
		}
		klog.Infof("  ✅ %s: proof bundle checkpoint(@%d) is consistent with current view of log(@%d)", p.name, fwCP.Size, lst.LatestConsistent.Size)
	}

	return errors.Join(errs...)
}

// extractFirmware attempts to read bootloader, os, and applet firmware and
// corresponding proof bundle information from the device.
func extractFirmware(dev string) (*firmwares, error) {
	f, err := os.OpenFile(dev, os.O_RDONLY, 0o400)
	if err != nil {
		return nil, fmt.Errorf("error opening %v: %v", dev, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Errorf("Error closing %v: %v", dev, err)
		}
	}()
	bootloader, err := readFirmware(f, bootloaderConfigBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to read bootloader IMX: %v", err)
	}
	os, err := readFirmware(f, osBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to read OS: %v", err)
	}
	applet, err := readFirmware(f, appletBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to read Applet: %v", err)
	}
	fw := &firmwares{
		Bootloader:    bootloader,
		TrustedApplet: applet,
		TrustedOS:     os,
	}
	return fw, nil
}

// readFirmware reads the config block from the MMC starting at the indicated block,
// and uses the information it contains to locate and read the corresponding firmware
// image too.
//
// A firmware bundle is returned with image + proof bundle information.
func readFirmware(f *os.File, cfgBlock int64) (firmware.Bundle, error) {
	cfg, err := readConfig(f, cfgBlock)
	if err != nil {
		return firmware.Bundle{}, err
	}

	fw := firmware.Bundle{
		Checkpoint:     cfg.Bundle.Checkpoint,
		Index:          cfg.Bundle.LogIndex,
		InclusionProof: cfg.Bundle.InclusionProof,
		Manifest:       cfg.Bundle.Manifest,
		Firmware:       make([]byte, cfg.Size),
	}
	klog.Infof("Found config at block 0x%x", cfgBlock)
	if klog.V(1).Enabled() {
		pp, _ := json.MarshalIndent(cfg, "", "  ")
		klog.V(1).Infof("Config:\n%s", pp)
	}
	klog.Infof("Reading 0x%x bytes of firmware from MMC byte offset 0x%x", cfg.Size, cfg.Offset)
	if _, err := f.ReadAt(fw.Firmware, cfg.Offset); err != nil {
		return firmware.Bundle{}, fmt.Errorf("failed to read firmware data: %v", err)
	}
	return fw, nil
}

// readConfig attempts to read and parse a GOB encoded Config structure from the
// MMC starting at the specificed block.
//
// This structure holds the location on MMC at which the corresponding firmware
// image can be found, as well as the FT proofbundle for that image.
func readConfig(f *os.File, cfgBlock int64) (*config.Config, error) {
	buf := make([]byte, config.MaxLength)
	if _, err := f.ReadAt(buf, cfgBlock*mmcBlockSize); err != nil {
		return nil, fmt.Errorf("failed to read config region @ block %d: %v", cfgBlock, err)
	}

	cfg := &config.Config{}
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %v", err)
	}

	return cfg, nil
}

// verifierFromFlags creates a new verifier from information passed in through flags.
func verifierFromFlags() verifier {
	var err error
	v := verifier{
		logOrigin: *firmwareLogOrigin,
	}
	v.logV, err = note.NewVerifier(*firmwareLogVerifier)
	if err != nil {
		klog.Exitf("Invalid firmware log verifier: %v", err)
	}
	v.appletV, err = note.NewVerifier(*appletVerifier)
	if err != nil {
		klog.Exitf("Invalid applet verifier: %v", err)
	}
	v.bootV, err = note.NewVerifier(*bootVerifier)
	if err != nil {
		klog.Exitf("Invalid boot verifier: %v", err)
	}
	v.osV1, err = note.NewVerifier(*osVerifier1)
	if err != nil {
		klog.Exitf("Invalid OS verifier 1: %v", err)
	}
	v.osV2, err = note.NewVerifier(*osVerifier2)
	if err != nil {
		klog.Exitf("Invalid OS verifier 2: %v", err)
	}
	v.recoveryV, err = note.NewVerifier(*recoveryVerifier)
	if err != nil {
		klog.Exitf("Invalid recovery verifier: %v", err)
	}

	v.logBaseURL, err = url.Parse(*firmwareLogURL)
	if err != nil {
		klog.Exitf("Firmware log URL invalid: %v", err)
	}
	v.binBaseURL, err = url.Parse(*binariesURL)
	if err != nil {
		klog.Exitf("Binaries URL invalid: %v", err)
	}

	return v
}
