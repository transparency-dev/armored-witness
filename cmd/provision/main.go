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
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/flynn/u2f/u2fhid"
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"

	"github.com/transparency-dev/armored-witness-boot/config"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"github.com/transparency-dev/armored-witness-common/release/firmware/update"
	"golang.org/x/mod/sumdb/note"
)

const (
	// Block size in bytes of the MMC device on the armored witness.
	mmcBlockSize = 512

	// bootloaderBlock defines the location of the first block of the bootloader on MMC.
	bootloaderBlock = 0x2
	// osBlock defines the location of the first block of the TrustedOS on MMC.
	osBlock = 0x5000
	// appletBlock defines the location of the first block of the TrustedApplet on MMC.
	appletBlock = 0x200000
)

var (
	recoveryImagePath = flag.String("recovery_image", "../armory-ums/armory-ums.imx", "Location of the recovery imx file.")
	bootloaderPath    = flag.String("bootloader", "../armored-witness-boot/armored-witness-boot.imx", "Location of the bootloader imx file.")
	trustedAppletPath = flag.String("trusted_applet", "../armored-witness-applet/bin/trusted_applet.elf", "Location of the trusted applet ELF file.")
	trustedOSPath     = flag.String("trusted_os", "../armored-witness-os/bin/trusted_os.elf", "Location of the trusted OS ELF file.")

	firmwareLogURL      = flag.String("firmware_log_url", "", "URL of the firmware transparency log to scan for firmware artefacts.")
	firmwareLogOrigin   = flag.String("firmware_log_origin", "", "Origin string for the firmware transparency log.")
	firmwareLogVerifier = flag.String("firmware_log_verifier", "", "Checkpoint verifier key for the firmware transparency log.")

	blockDeviceGlob = flag.String("blockdevs", "/dev/sd*", "Glob for plausible block devices where the armored witness could appear")

	runAnyway = flag.Bool("run_anyway", false, "Let the user override bailing on any potential problems we've detected.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
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

// firmware respresents the collection of firmware and related artefacts which must be
// flashed onto the device.
type firmware struct {
	// Bootloader holds the regular bootloader as an unsigned IMX.
	Bootloader []byte
	// BootloaderBlock is the location on MMC where the bootloader should be written.
	BootloaderBlock int64
	// Recovery holds the recovery-boot image as an unsigned IMX.
	Recovery []byte
	// TrustedOS holds the trusted OS firmware as a signed ELF.
	TrustedOS     []byte
	TrustedOSSig1 []byte
	TrustedOSSig2 []byte
	// TrustedOSBlock is the location on MMC where the TrustedOS should be written.
	TrustedOSBlock int64
	// TrustedApplet holds the witness applet firmware as a signed ELF.
	TrustedApplet    []byte
	TrustedAppletSig []byte
	// TrustedAppletBlock is the location on MMC where the TrustedApplet should be written.
	TrustedAppletBlock int64
	// TODO: add proof bundles, etc.
}

func fetchLatestArtefacts(ctx context.Context) (*firmware, error) {
	logBaseURL, err := url.Parse(*firmwareLogURL)
	if err != nil {
		return nil, fmt.Errorf("firmware log URL invalid: %v", err)
	}

	logVerifier, err := note.NewVerifier(*firmwareLogVerifier)
	if err != nil {
		return nil, fmt.Errorf("invalid firmware log verifier: %v", err)
	}
	binFetcher := func(_ context.Context, r ftlog.FirmwareRelease) ([]byte, error) {
		klog.Infof("Asked to get bin for %v", r)
		return nil, errors.New("not implemented")
	}

	updateFetcher, err := update.NewFetcher(ctx,
		update.FetcherOpts{
			LogFetcher:    newLogFetcher(logBaseURL),
			LogOrigin:     *firmwareLogOrigin,
			LogVerifier:   logVerifier,
			BinaryFetcher: binFetcher,
		})
	if err != nil {
		return nil, fmt.Errorf("NewFetcher: %v", err)
	}

	if err := updateFetcher.Scan(ctx); err != nil {
		return nil, fmt.Errorf("Scan: %v", err)
	}

	latestOsVer, latestAppletVer, err := updateFetcher.GetLatestVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetLatestVersions: %v", err)
	}

	klog.Infof("Found latest versions: OS %v, Applet %v", latestOsVer, latestAppletVer)

	osMeta, err := updateFetcher.GetOS(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetOS: %v", err)
	}
	klog.Infof("OS bundle:\n%v", osMeta)

	appletMeta, err := updateFetcher.GetApplet(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetOS: %v", err)
	}
	klog.Infof("Applet bundle:\n%v", appletMeta)

	// TODO: Use the armored witness transparency logs as the source of firmware images.
	// For now, we'll just expect that repos are checked-out in adjacent directories,
	// and images have been built there.
	fw := &firmware{
		BootloaderBlock:    bootloaderBlock,
		TrustedOSBlock:     osBlock,
		TrustedAppletBlock: appletBlock,
	}

	if fw.Recovery, err = os.ReadFile(*recoveryImagePath); err != nil {
		return nil, fmt.Errorf("failed to read recovery image from %q: %v", *recoveryImagePath, err)
	}
	if fw.Bootloader, err = os.ReadFile(*bootloaderPath); err != nil {
		return nil, fmt.Errorf("failed to read bootloader from %q: %v", *bootloaderPath, err)
	}
	if fw.TrustedApplet, err = os.ReadFile(*trustedAppletPath); err != nil {
		return nil, fmt.Errorf("failed to read trusted applet from %q: %v", *trustedAppletPath, err)
	}
	appletSigPath := strings.ReplaceAll(*trustedAppletPath, ".elf", ".sig")
	if fw.TrustedAppletSig, err = os.ReadFile(appletSigPath); err != nil {
		return nil, fmt.Errorf("failed to read trusted applet signature from %q: %v", appletSigPath, err)
	}
	if fw.TrustedOS, err = os.ReadFile(*trustedOSPath); err != nil {
		return nil, fmt.Errorf("failed to read trusted OS from %q: %v", *trustedOSPath, err)
	}
	osSig1Path := strings.ReplaceAll(*trustedOSPath, ".elf", ".sig1")
	if fw.TrustedOSSig1, err = os.ReadFile(osSig1Path); err != nil {
		return nil, fmt.Errorf("failed to read trusted OS sig from %q: %v", osSig1Path, err)
	}
	osSig2Path := strings.ReplaceAll(*trustedOSPath, ".elf", ".sig2")
	if fw.TrustedOSSig2, err = os.ReadFile(osSig2Path); err != nil {
		return nil, fmt.Errorf("failed to read trusted OS sig from %q: %v", osSig2Path, err)
	}

	klog.Info("Loaded firmware artefacts:")
	klog.Infof("Recovery:      SHA256:%032x (%s)", sha256.Sum256(fw.Recovery), *recoveryImagePath)
	klog.Infof("Bootloader:    SHA256:%032x (%s)", sha256.Sum256(fw.Bootloader), *bootloaderPath)
	klog.Infof("TrustedApplet: SHA256:%032x (%s)", sha256.Sum256(fw.TrustedApplet), *trustedAppletPath)
	klog.Infof("TrustedOS:     SHA256:%032x (%s)", sha256.Sum256(fw.TrustedOS), *trustedOSPath)

	// OS and Applet need prepending with config structures.
	if fw.TrustedApplet, err = prepareELF(fw.TrustedApplet, [][]byte{fw.TrustedAppletSig}, appletBlock); err != nil {
		return nil, fmt.Errorf("failed to prepare TrustedApplet: %v", err)
	}
	if fw.TrustedOS, err = prepareELF(fw.TrustedOS, [][]byte{fw.TrustedOSSig1, fw.TrustedOSSig2}, osBlock); err != nil {
		return nil, fmt.Errorf("failed to prepare TrustedOS: %v", err)
	}

	return fw, nil
}

// waitAndProvision waits for a fresh armored witness device to be detected, and then provisions it.
func waitAndProvision(ctx context.Context, fw *firmware) error {
	klog.Info("Operator, please ensure boot switch is set to USB, and then connect unprovisioned device 🙏")
	// The device will initially be in HID mode (showing as "RecoveryMode" in the output to lsusb).
	// So we'll detect it as such:
	target, err := waitForHIDDevice(ctx)
	if err != nil {
		return err
	}
	klog.Infof("✅ Detected device %q", target.DeviceInfo.Path)

	_, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate ephemeral key: %v", err)
	}

	// Per-device prep:
	// TODO: sign bootloader and recovery images.
	// TODO: store signed bootloader and recovery images somewhere durable.

	// SDP boot recovery image on device.
	// Booting the recovery image causes the device re-appear as a USB Mass Storage device.
	// So we'll wait for that to happen, and figure out which /dev/ entry corresponds to it.
	bDev, err := waitForBlockDevice(ctx, *blockDeviceGlob, func() error {
		if err := target.BootIMX(fw.Recovery); err != nil {
			return fmt.Errorf("failed to SDP boot recovery image on %v: %v", target.DeviceInfo.Path, err)
		}
		klog.Info("✅ Witness device booting recovering image")
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to detect block device: %v", err)

	}
	klog.Infof("✅ Detected blockdevice %v", bDev)

	for i := 5; i > 0; i-- {
		klog.Infof("  Flashing in %d", i)
		<-time.After(time.Second)
	}

	klog.Infof("Flashing images...")
	if err := flashImages(bDev, fw); err != nil {
		return fmt.Errorf("error while flashing images: %v", err)
	}
	klog.Info("✅ Flashed all images")

	// TODO: Write proof bundle.

	klog.Info("Operator, please change boot switch to MMC, and then reboot device 🙏")
	klog.Info("Waiting for device to boot...")

	p, dev, err := waitForU2FDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to find armored witness device: %v", err)
	}
	defer dev.Close()

	klog.Infof("✅ Detected device %q", p)
	s, err := witnessStatus(dev)
	if err != nil {
		return fmt.Errorf("failed to fetch witness status: %v", err)
	}

	klog.Infof("✅ Witness serial number %s found", s.Serial)
	if s.HAB {
		return fmt.Errorf("witness serial number %s has HAB fuse set!", s.Serial)
	}
	klog.Infof("✅ Witness serial number %s is not HAB fused", s.Serial)

	// TODO: Set fuses.

	// TODO: Reboot device.

	// TODO: Use HID to access witness public keys from device and store somewhere durable.

	klog.Infof("✅ Witness ID %s provisioned", s.Witness.Identity)

	return nil

}

// waitForHIDDevice waits for an unprovisioned armored witness device
// to appear on the USB bus.
func waitForHIDDevice(ctx context.Context) (*Target, error) {
	klog.Info("Waiting for device to be detected...")
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			targets, err := DetectHID()
			if err != nil {
				klog.Warningf("Failed to detect devices: %v", err)
				continue
			}
			if len(targets) == 0 {
				continue
			}
			return targets[0], nil
		}
	}
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
			p, target, err := detectU2F()
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

// waitForBlockDevice runs f, and waits for a block device matching glob to appear.
func waitForBlockDevice(ctx context.Context, glob string, f func() error) (string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return "", fmt.Errorf("failed to create fs watcher: %v", err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			klog.Errorf("Error closing fs watcher: %v", err)
		}
	}()

	// Set up the watcher to look for events in /dev only.
	if err := watcher.Add("/dev"); err != nil {
		return "", fmt.Errorf("failed to add /dev to fs watcher: %v", err)
	}

	// Run the passed-in function
	if err := f(); err != nil {
		return "", err
	}

	// Finally, monitor fsnotify events for any which match the glob.
	klog.Info("Waiting for block device to appear")
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case e := <-watcher.Events:
			matched, err := filepath.Match(*blockDeviceGlob, e.Name)
			if err != nil {
				klog.Exitf("error testing filename %q against glob %q: %v", e.Name, *blockDeviceGlob, err)
			}
			if matched && e.Has(fsnotify.Create) {
				return e.Name, nil
			}
		}
	}
}

// flashImages writes all the images in fw to the specified block device.
func flashImages(dev string, fw *firmware) error {
	f, err := os.OpenFile(dev, os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("error opening %v: %v", dev, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Errorf("Errorf closing %v: %v", dev, err)
		}
	}()

	for _, p := range []struct {
		name  string
		img   []byte
		block int64
	}{
		{name: "Bootloader", img: fw.Bootloader, block: fw.BootloaderBlock},
		{name: "TrustedOS", img: fw.TrustedOS, block: fw.TrustedOSBlock},
		{name: "TrustedApplet", img: fw.TrustedApplet, block: fw.TrustedAppletBlock},
	} {
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
func prepareELF(elf []byte, sigs [][]byte, block int64) ([]byte, error) {
	conf := &config.Config{
		Offset:     block*mmcBlockSize + config.MaxLength,
		Size:       int64(len(elf)),
		Signatures: sigs,
	}

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(conf); err != nil {
		return nil, fmt.Errorf("failed to encode config: %v", err)
	}

	pad := config.MaxLength - int64(buf.Len())
	buf.Write(make([]byte, pad))
	buf.Write(elf)

	return buf.Bytes(), nil
}
