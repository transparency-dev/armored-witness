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

// provision is a tool for helping with the initial configuration and
// flashing of ArmoredWitness devices.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"time"

	"github.com/flynn/u2f/u2fhid"
	"k8s.io/klog/v2"

	"github.com/transparency-dev/armored-witness-boot/config"
	"github.com/transparency-dev/armored-witness-common/release/firmware"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"github.com/transparency-dev/armored-witness-common/release/firmware/update"
	"github.com/transparency-dev/armored-witness/internal/device"
	"github.com/transparency-dev/armored-witness/internal/fetcher"
	"github.com/transparency-dev/armored-witness/internal/release"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/sumdb/note"
)

const (
	// Block size in bytes of the MMC device on the armored witness.
	mmcBlockSize = 512

	// bootloaderBlock defines the location of the first block of the bootloader on MMC.
	bootloaderBlock = 0x2
	// bootloaderConfigBlock defines the location of the bootloader config GOB on MMC.
	// In constrast to the other firmware binaries below where each firmware is preceeded
	// by its config GOB, the bootloader config is stored separatly due to the hard requirement
	// for the binary location imposed by the i.MX ROM bootloader.
	bootloaderConfigBlock = 0x4FB0
	// osBlock defines the location of the first block of the TrustedOS on MMC.
	osBlock = 0x5000
	// appletBlock defines the location of the first block of the TrustedApplet on MMC.
	appletBlock = 0x200000
	// appletDataBlock defines the location of the applet data storage area.
	appletDataBlock = 0x400000
	// appletDataNumBlocks is the number of blocks in the applet data storage area.
	appletDataNumBlocks = 0x400000

	fuseWarning = `
████████████████████████████████████████████████████████████████████████████████

                                **  WARNING  **

Enabling NXP HABv4 secure boot is an irreversible action that permanently fuses
verification key hashes on the device.

Any errors in the process or loss of the signing PKI will result in a bricked
device incapable of executing unsigned code. This is a security feature, not a
bug.

The use of this tool is therefore **at your own risk**.

████████████████████████████████████████████████████████████████████████████████
`

	operPlease = "🔷🔷🔷 🙋 OPERATOR: %s 🙏"
)

var (
	// expectedSRKHashes maps known SRK hash values to the release environment they came from.
	// These values MUST NOT be changed unless you really know what you're doing!
	expectedSRKHashes = map[string]string{
		// ci: From https://github.com/transparency-dev/armored-witness/blob/main/deployment/build_and_release/live/ci/terragrunt.hcl#L32
		"b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a": "ci",
		// prod: From https://github.com/transparency-dev/armored-witness/blob/main/deployment/build_and_release/live/prod/terragrunt.hcl#L33
		"77e021cc51b5547fb0c2192fb32710bfa89b4bbaa7dab5f97fc585f673b0b236": "prod",
	}
)

var (
	template            = flag.String("template", "", fmt.Sprintf("One of the optional preconfigured templates (%v)", maps.Keys(release.Templates)))
	firmwareLogURL      = flag.String("firmware_log_url", "", "URL of the firmware transparency log to scan for firmware artefacts.")
	firmwareLogOrigin   = flag.String("firmware_log_origin", "", "Origin string for the firmware transparency log.")
	firmwareLogVerifier = flag.String("firmware_log_verifier", "", "Checkpoint verifier key for the firmware transparency log.")
	binariesURL         = flag.String("binaries_url", "", "Base URL for fetching firmware artefacts referenced by FT log.")

	appletVerifier   = flag.String("applet_verifier", "", "Verifier key for the applet manifest.")
	bootVerifier     = flag.String("boot_verifier", "", "Verifier key for the boot manifest.")
	osVerifier1      = flag.String("os_verifier_1", "", "Verifier key 1 for the OS manifest.")
	osVerifier2      = flag.String("os_verifier_2", "", "Verifier key 2 for the OS manifest.")
	recoveryVerifier = flag.String("recovery_verifier", "", "Verifier key for the recovery manifest.")

	osIndexOverride     = flag.Int64("os_index", -1, "Override the OS to install by specifying the index into the log for its manifest.")
	appletIndexOverride = flag.Int64("applet_index", -1, "Override the Applet to install by specifying the index into the log for its manifest.")

	habTarget       = flag.String("hab_target", "", "Device type firmware must be targetting.")
	blockDeviceGlob = flag.String("blockdevs", "/dev/disk/by-id/usb-F-Secure_USB_*0:0", "Glob for plausible block devices where the armored witness could appear.")

	runAnyway   = flag.Bool("run_anyway", false, "Let the user override bailing on any potential problems we've detected.")
	wipeWitness = flag.Bool("wipe_witness_state", false, "If true, erase the witness stored data.")

	fuse = flag.Bool("fuse", false, "If set, device will be **permanently** fused to the release environment specified by --hab_target")
)

func applyFlagTemplate(k string) {
	t, ok := release.Templates[k]
	if !ok {
		klog.Exitf("No such template %q", k)
	}
	for f, v := range t {
		if u := flag.Lookup(f); u == nil {
			klog.Exitf("Internal error - template flag --%v unknown", f)
		} else if u.Value.String() != "" {
			klog.Exitf("Cannot set --template and --%s", f)
		}
		klog.Infof("Using template flag setting --%v=%v", f, v)
		if err := flag.Set(f, v); err != nil {
			klog.Exitf("Failed to set template flag --%v: %v", f, err)
		}
	}
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	if *template != "" {
		applyFlagTemplate(*template)
	}
	ctx := context.Background()

	if u, err := user.Current(); err != nil {
		klog.Exitf("Failed to determine who I'm running as: %v", err)
	} else if u.Uid != "0" {
		klog.Warningf("⚠️ This tool probably needs to be run as root (e.g. via sudo), it's running as %q (UID %q); re-run with the --run_anyway flag if you know better.", u.Username, u.Uid)
		if !*runAnyway {
			klog.Exit("Bailing.")
		}
	}

	fw, err := fetchLatestArtefacts(ctx)
	if err != nil {
		klog.Exitf("Failed to fetch latest firmware artefacts: %v", err)
	}

	if err := waitAndProvision(ctx, fw); err != nil {
		klog.Exitf("❌ Failed to provision device: %v", err)
	}
	klog.Info("✅ Device provisioned!")
}

// Firmware represents a single firmware image to be be installed on a device.
type fw struct {
	// bundle is the firmware bundle to be installed
	bundle firmware.Bundle
	// block is the location on MMC where the firmware should be installed.
	block int64
	// configBlock is an optional location on the MMC where the loader config block should be
	// stored. This is only currently used for the bootloader firmware due to location constraints.
	configBlock int64
}

// firmwares respresents the collection of firmware and related artefacts which must be
// flashed onto the device.
type firmwares struct {
	// bootloader holds the regular bootloader firmware bundle.
	bootloader *fw
	// recovery holds the recovery-boot image as an unsigned IMX.
	recovery *fw
	// trustedOS holds the trusted OS firmware bundle.
	trustedOS *fw
	// trustedApplet holds the witness applet firmware bundle.
	trustedApplet *fw
}

type firmwareJobs struct {
	bootloader       flashJob
	bootloaderConfig flashJob
	trustedOS        flashJob
	trustedApplet    flashJob
}

type bundleProvider interface {
	GetOS(ctx context.Context) (firmware.Bundle, error)
	GetApplet(ctx context.Context) (firmware.Bundle, error)
	GetRecovery(ctx context.Context) (firmware.Bundle, error)
	GetBoot(ctx context.Context) (firmware.Bundle, error)
}

func fetchLatestArtefacts(ctx context.Context) (*firmwares, error) {
	logBaseURL, err := url.Parse(*firmwareLogURL)
	if err != nil {
		return nil, fmt.Errorf("firmware log URL invalid: %v", err)
	}

	logVerifier, err := note.NewVerifier(*firmwareLogVerifier)
	if err != nil {
		return nil, fmt.Errorf("invalid firmware log verifier: %v", err)
	}
	appletVerifier, err := note.NewVerifier(*appletVerifier)
	if err != nil {
		return nil, fmt.Errorf("invalid applet verifier: %v", err)
	}
	bootVerifier, err := note.NewVerifier(*bootVerifier)
	if err != nil {
		return nil, fmt.Errorf("invalid boot verifier: %v", err)
	}
	osVerifier1, err := note.NewVerifier(*osVerifier1)
	if err != nil {
		return nil, fmt.Errorf("invalid OS verifier 1: %v", err)
	}
	osVerifier2, err := note.NewVerifier(*osVerifier2)
	if err != nil {
		return nil, fmt.Errorf("invalid OS verifier 2: %v", err)
	}
	recoveryVerifier, err := note.NewVerifier(*recoveryVerifier)
	if err != nil {
		return nil, fmt.Errorf("invalid recovery verifier: %v", err)
	}

	binBaseURL, err := url.Parse(*binariesURL)
	if err != nil {
		return nil, fmt.Errorf("binaries URL invalid: %v", err)
	}
	binFetcher := fetcher.BinaryFetcher(fetcher.New(binBaseURL))

	updateFetcher, err := update.NewFetcher(ctx,
		update.FetcherOpts{
			LogFetcher:       fetcher.New(logBaseURL),
			LogOrigin:        *firmwareLogOrigin,
			LogVerifier:      logVerifier,
			BinaryFetcher:    binFetcher,
			AppletVerifier:   appletVerifier,
			BootVerifier:     bootVerifier,
			OSVerifiers:      [2]note.Verifier{osVerifier1, osVerifier2},
			RecoveryVerifier: recoveryVerifier,
			HABTarget:        *habTarget,
		})
	if err != nil {
		return nil, fmt.Errorf("NewFetcher: %v", err)
	}

	if err := updateFetcher.Scan(ctx); err != nil {
		return nil, fmt.Errorf("updateFetcher.Scan: %v", err)
	}

	fetchSession, err := updateFetcher.NewSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("updateFetcher.NewSession: %v", err)
	}
	bp := overridableBundleProvider{
		delegate:     updateFetcher,
		binFetcher:   binFetcher,
		fetchSession: fetchSession,
	}

	firmwares := &firmwares{}

	osFW, err := bp.GetOS(ctx)
	if err != nil {
		return nil, fmt.Errorf("bp.GetOS: %v", err)
	}
	klog.Infof("Found OS bundle @ %d", osFW.Index)
	firmwares.trustedOS = &fw{
		bundle: osFW,
		block:  osBlock,
	}

	appletFW, err := bp.GetApplet(ctx)
	if err != nil {
		return nil, fmt.Errorf("bp.GetApplet: %v", err)
	}
	klog.Infof("Found Applet bundle @ %d", appletFW.Index)
	firmwares.trustedApplet = &fw{
		bundle: appletFW,
		block:  appletBlock,
	}

	bootFW, err := bp.GetBoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("bp.GetBoot: %v", err)
	}
	klog.Infof("Found Bootloader bundle @ %d", bootFW.Index)
	firmwares.bootloader = &fw{
		bundle:      bootFW,
		block:       bootloaderBlock,
		configBlock: bootloaderConfigBlock,
	}

	recoveryFW, err := bp.GetRecovery(ctx)
	if err != nil {
		return nil, fmt.Errorf("bp.GetRecovery: %v", err)
	}
	klog.Infof("Found Recovery bundle @ %d", recoveryFW.Index)
	firmwares.recovery = &fw{
		bundle: recoveryFW,
	}

	klog.Info("Loaded firmware artefacts.")
	return firmwares, nil
}

func prepareFlashJobs(firmwares *firmwares) (*firmwareJobs, error) {
	jobs := &firmwareJobs{}
	if firmwares.trustedOS != nil {
		// OS and Applet need prepending with config structures.
		osAndConfig, err := prepareELF(firmwares.trustedOS.bundle, firmwares.trustedOS.block)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare TrustedOS: %v", err)
		}
		jobs.trustedOS = flashJob{name: "os", img: osAndConfig, block: firmwares.trustedOS.block}
	}
	if firmwares.trustedApplet != nil {
		appletAndConfig, err := prepareELF(firmwares.trustedApplet.bundle, firmwares.trustedApplet.block)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare TrustedApplet: %v", err)
		}
		jobs.trustedApplet = flashJob{name: "applet", img: appletAndConfig, block: firmwares.trustedApplet.block}
	}
	if firmwares.bootloader != nil {
		bootloaderConfig, err := configFromBundle(firmwares.bootloader.bundle, firmwares.bootloader.block*mmcBlockSize)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare Bootloader config: %v", err)
		}
		jobs.bootloaderConfig = flashJob{name: "boot config", img: bootloaderConfig, block: firmwares.bootloader.configBlock}

		bootloaderHAB := append(firmwares.bootloader.bundle.Firmware, firmwares.bootloader.bundle.HABSignature...)
		klog.Infof("Bootloader firmware is %d bytes + %d bytes HAB signature", len(firmwares.bootloader.bundle.Firmware), len(firmwares.bootloader.bundle.HABSignature))
		jobs.bootloader = flashJob{name: "bootloader", img: bootloaderHAB, block: firmwares.bootloader.block}
	}
	return jobs, nil
}

// waitAndProvision waits for a fresh armored witness device to be detected, and then provisions it.
func waitAndProvision(ctx context.Context, fw *firmwares) error {
	klog.Infof(operPlease, "please ensure boot switch is set to USB (towards RJ45 socket), and then connect unprovisioned device")

	recoveryHAB := append(fw.recovery.bundle.Firmware, fw.recovery.bundle.HABSignature...)
	klog.Infof("Recovery firmware is %d bytes + %d bytes HAB signature", len(fw.recovery.bundle.Firmware), len(fw.recovery.bundle.HABSignature))

	// The device will initially be in HID mode (showing as "RecoveryMode" in the output to lsusb).
	// So we'll detect it as such:
	target, bDev, err := device.BootIntoRecovery(ctx, recoveryHAB, *blockDeviceGlob)
	if err != nil {
		return err
	}
	klog.Infof("✅ Detected device %q", target.DeviceInfo.Path)
	klog.Infof("✅ Detected blockdevice %v", bDev)

	_, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate ephemeral key: %v", err)
	}

	jobs, err := prepareFlashJobs(fw)
	if err != nil {
		return fmt.Errorf("failed to prepare flash jobs: %v", err)
	}

	flashStages := [][]flashJob{{jobs.trustedOS, jobs.bootloader, jobs.bootloaderConfig}}
	if *fuse {
		// If we need to fuse the device, we'll install the applet later on.
		// This is ensure there's no unexpected CPU load on the device when
		// we attempt to set fuses as this has been known to be timing sensitive.

		// Add an extra job into the first stage to destroy any pre-existing applet installed on the device.
		flashStages[0] = append(flashStages[0], flashJob{name: "DUMMY_APPLET", img: make([]byte, len(jobs.trustedApplet.img)), block: jobs.trustedApplet.block})
		// Then create a second stage job to install the real applet
		flashStages = append(flashStages, []flashJob{jobs.trustedApplet})
	} else {
		// We're not fusing, so can just install everything in the first stage.
		flashStages[0] = append(flashStages[0], jobs.trustedApplet)

	}
	klog.Infof("Flashing images...")
	if err := flashImages(bDev, flashStages[0]); err != nil {
		return fmt.Errorf("error while flashing images: %v", err)
	}
	klog.Info("✅ Flashed images")

	if *wipeWitness {
		if err := wipeAppletData(bDev); err != nil {
			return fmt.Errorf("error while wiping applet data: %v", err)
		}
	}

	klog.Infof(operPlease, "please change boot switch to MMC (away from RJ45 socket), and then reboot device")
	klog.Info("Waiting for device to boot...")

	p, dev, err := waitForU2FDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to find armored witness device: %v", err)
	}

	klog.Infof("✅ Detected device %q", p)
	s, err := device.WitnessStatus(dev)
	if err != nil {
		return fmt.Errorf("failed to fetch witness status: %v", err)
	}
	klog.Infof("✅ Witness serial number %s found", s.Serial)
	if s.HAB {
		if *fuse && !*runAnyway {
			return fmt.Errorf("witness serial number %s has HAB fuse set", s.Serial)
		}
		klog.Infof("⚠️  Witness serial number %s is already HAB fused", s.Serial)
	} else {
		klog.Infof("✅ Witness serial number %s is not HAB fused", s.Serial)
	}

	srkEnv, ok := expectedSRKHashes[s.SRKHash]
	if !ok {
		e := fmt.Errorf("witness OS reports UNKNOWN SRK Hash '%s', not fusing", s.SRKHash)
		if *fuse {
			return e
		}
		klog.Warningf("⚠️  %s", e.Error())
	}
	if srkEnv != *habTarget {
		e := fmt.Errorf("witness OS reports SRK Hash (%s) for unexpected release environment %q - we're set to %q, not fusing", s.SRKHash, srkEnv, *habTarget)
		if *fuse {
			return e
		}
		klog.Warningf("⚠️  %s", e.Error())
	}

	if *fuse {
		klog.Warningf("\n%s\n", fuseWarning)
		for i := 5; i > 0; i-- {
			klog.Infof(" Fusing in %d", i)
			<-time.After(time.Second)
		}
		klog.Infof("Attempting to fuse device and activate HAB 🫣")
		if err := device.ActivateHAB(dev); err != nil {
			err = fmt.Errorf("device failed to activate HAB: %v", err)
			if !*runAnyway {
				return err
			}
			klog.Warningf("⚠️  %s, continuing anyway", err.Error())
		}
		klog.Info("✅ Fusing successful! 👌")
		// Close dev as we'll need to re-open it below after the device has rebooted...
		dev.Close()

		klog.Infof("%d remaining firmware(s) to install", len(flashStages[1]))

		klog.Infof(operPlease, "please change boot switch to USB (towards RJ45 socket), and then reboot device")
		klog.Info("Waiting for device to boot...")
		// The device will initially be in HID mode (showing as "RecoveryMode" in the output to lsusb).
		// So we'll detect it as such:
		target, bDev, err := device.BootIntoRecovery(ctx, recoveryHAB, *blockDeviceGlob)
		if err != nil {
			return err
		}
		klog.Infof("✅ Detected device %q", target.DeviceInfo.Path)
		klog.Infof("✅ Detected blockdevice %v", bDev)

		klog.Infof("Flashing Applet image...")
		if err := flashImages(bDev, flashStages[1]); err != nil {
			return fmt.Errorf("error while flashing Applet image: %v", err)
		}
		klog.Info("✅ Flashed Applet image")

		klog.Infof(operPlease, "please change boot switch to MMC (away from RJ45 socket), and then reboot device")
		klog.Info("Waiting for device to boot...")
		p, dev, err = waitForU2FDevice(ctx)
		if err != nil {
			return fmt.Errorf("failed to find armored witness device: %v", err)
		}

		klog.Infof("✅ Detected device %q", p)
		s, err = device.WitnessStatus(dev)
		if err != nil {
			return fmt.Errorf("failed to fetch witness status: %v", err)
		}
	}

	klog.Infof(operPlease, "please reboot device")
	klog.Info("Waiting for device to boot...")

	klog.Infof("✅ Witness ID %s provisioned", s.Witness.Identity)

	dev.Close()

	return nil

}

// waitForU2FDevice waits for a device running armored witness firmware
// to appear on the USB bus.
// Returns the device path & opened device.
func waitForU2FDevice(ctx context.Context) (string, *u2fhid.Device, error) {
	klog.Info("Waiting for armored witness device to be detected...")
	for {
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(time.Second):
			p, target, err := device.DetectU2F()
			if err != nil {
				klog.Warningf("Failed to detect devices: %v", err)
				continue
			}
			if target == nil {
				continue
			}
			return p, target, nil
		}
	}
}

type flashJob struct {
	name  string
	img   []byte
	block int64
}

// flashImages writes all the images in fw to the specified block device.
func flashImages(dev string, jobs []flashJob) error {
	for i := 5; i > 0; i-- {
		klog.Infof("  Flashing in %d", i)
		<-time.After(time.Second)
	}

	f, err := os.OpenFile(dev, os.O_RDWR|os.O_SYNC, 0o600)
	if err != nil {
		return fmt.Errorf("error opening %v: %v", dev, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Errorf("Error closing %v: %v", dev, err)
		}
	}()

	for _, p := range jobs {
		if err := flashImage(p.img, f, p.block); err != nil {
			klog.Infof("  ❌ %s", p.name)
			return fmt.Errorf("failed to flash %s: %v", p.name, err)
		}
		klog.Infof("  ✅ %s @ 0x%0x", p.name, p.block)
	}
	return nil
}

// flashImage writes the image to the file starting at the specified block.
func flashImage(image []byte, to *os.File, atBlock int64) error {
	offset := atBlock * mmcBlockSize
	if n, err := to.WriteAt(image, offset); err != nil {
		return err
	} else if l := len(image); n != l {
		return fmt.Errorf("short write (%d < %d)", n, l)
	}
	return to.Sync()
}

// prepareELF returns a slice with a GOB-encoded configuration structure followed by the ELF image.
// block is the first MMC block of the config+ELF region.
func prepareELF(bundle firmware.Bundle, block int64) ([]byte, error) {
	// For ELF firmwares (OS & Applet), the on-MMC layout is [configGOB|padding|ELF]
	fwOffset := block*mmcBlockSize + config.MaxLength
	cfgGob, err := configFromBundle(bundle, fwOffset)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	buf.Write(cfgGob)
	pad := config.MaxLength - int64(buf.Len())
	buf.Write(make([]byte, pad))
	buf.Write(bundle.Firmware)

	return buf.Bytes(), nil
}

// configFromBundle creates a populated config struct using the passed in contents.
// firmwareOffset is the offset in bytes of the firmware binary from the start of the MMC.
func configFromBundle(bundle firmware.Bundle, firmwareOffset int64) ([]byte, error) {
	conf := &config.Config{
		Offset: firmwareOffset,
		Size:   int64(len(bundle.Firmware)),
		Bundle: config.ProofBundle{
			Checkpoint:     bundle.Checkpoint,
			Manifest:       bundle.Manifest,
			LogIndex:       bundle.Index,
			InclusionProof: bundle.InclusionProof,
		},
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(conf); err != nil {
		return nil, fmt.Errorf("failed to encode config: %v", err)
	}

	return buf.Bytes(), nil
}

// wipeAppletData erases MMC blocks allocated to applet data storage.
func wipeAppletData(dev string) error {
	f, err := os.OpenFile(dev, os.O_RDWR|os.O_SYNC, 0o600)
	if err != nil {
		return fmt.Errorf("error opening %v: %v", dev, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Errorf("Errorf closing %v: %v", dev, err)
		}
	}()

	klog.Infof("Wiping data area blocks [0x%x, 0x%x)...", appletDataBlock, appletDataBlock+appletDataNumBlocks)
	chunkBlocks := 2048
	empty := make([]byte, chunkBlocks*mmcBlockSize)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for i := 0; i < appletDataNumBlocks; i += chunkBlocks {
		select {
		case <-t.C:
			klog.Infof("   %3d%%", (i*100)/appletDataNumBlocks)
		default:
		}

		offset := (int64(appletDataBlock+i) * mmcBlockSize)
		if appletDataNumBlocks-i < chunkBlocks {
			chunkBlocks = appletDataNumBlocks - i
			empty = empty[:chunkBlocks*mmcBlockSize]
		}
		if _, err := f.WriteAt(empty, offset); err != nil {
			return fmt.Errorf("WriteAt: %v", err)
		}
	}
	klog.Info("   100%")
	return nil
}

type overridableBundleProvider struct {
	delegate bundleProvider

	fetchSession update.FetchSession
	binFetcher   update.BinaryFetcher
}

func (p *overridableBundleProvider) GetOS(ctx context.Context) (firmware.Bundle, error) {
	if *osIndexOverride < 0 {
		return p.delegate.GetOS(ctx)
	}

	i := uint64(*osIndexOverride)

	bundle, release, err := p.fetchSession.Fetch(ctx, i)
	if err != nil {
		return firmware.Bundle{}, err
	}
	if release.Component != ftlog.ComponentOS {
		return firmware.Bundle{}, fmt.Errorf("overridden OS index %d is of type %s, not required type %s", i, release.Component, ftlog.ComponentOS)
	}
	bundle.Firmware, _, err = p.binFetcher(ctx, *release)
	if err != nil {
		return firmware.Bundle{}, fmt.Errorf("binFetcher(): %v", err)
	}

	return *bundle, nil
}

func (p *overridableBundleProvider) GetApplet(ctx context.Context) (firmware.Bundle, error) {
	if *appletIndexOverride < 0 {
		return p.delegate.GetApplet(ctx)
	}

	i := uint64(*appletIndexOverride)

	bundle, release, err := p.fetchSession.Fetch(ctx, i)
	if err != nil {
		return firmware.Bundle{}, err
	}
	if release.Component != ftlog.ComponentApplet {
		return firmware.Bundle{}, fmt.Errorf("overridden Applet index %d is of type %s, not required type %s", i, release.Component, ftlog.ComponentApplet)
	}
	bundle.Firmware, _, err = p.binFetcher(ctx, *release)
	if err != nil {
		return firmware.Bundle{}, fmt.Errorf("binFetcher(): %v", err)
	}

	return *bundle, nil
}

func (p *overridableBundleProvider) GetRecovery(ctx context.Context) (firmware.Bundle, error) {
	return p.delegate.GetRecovery(ctx)
}

func (p *overridableBundleProvider) GetBoot(ctx context.Context) (firmware.Bundle, error) {
	return p.delegate.GetBoot(ctx)
}
