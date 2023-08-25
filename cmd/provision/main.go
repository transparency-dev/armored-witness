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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/klog"
)

var (
	recoveryImagePath = flag.String("recovery_image", "../armory-ums/armory-ums.imx", "Location of the recovery imx file.")
	bootloaderPath    = flag.String("bootloader", "../armored-witness-boot/armored-witness-boot.imx", "Location of the bootloader imx file.")
	trustedAppletPath = flag.String("trusted_applet", "../armored-witness-applet/bin/trusted_applet.elf", "Location of the trusted applet ELF file.")
	trustedOSPath     = flag.String("trusted_os", "../armored-witness-os/bin/trusted_os.elf", "Location of the trusted OS ELF file.")
)

func main() {
	if err := flag.Set("logtostderr", "true"); err != nil {
		klog.Exitf("Unable to set flag logtostderr to true: %v", err)
	}
	flag.Parse()

	fw, err := fetchLatestArtefacts()
	if err != nil {
		klog.Exitf("Failed to fetch latest firmware artefacts: %v", err)
	}

	if err := waitAndProvision(fw); err != nil {
		klog.Exitf("Failed to provision device: %v", err)
	}
}

type firmware struct {
	// BootLoader holds the regular bootloader as an unsigned IMX.
	BootLoader []byte
	// Recovery holds the recovery-boot image as an unsigned IMX.
	Recovery []byte
	// TrustedOS holds the trusted OS firmware as a signed ELF.
	TrustedOS []byte
	// TrustedApplet holds the witness applet firmware as a signed ELF.
	TrustedApplet []byte
	// TODO: add proof bundles, etc.
}

func fetchLatestArtefacts() (*firmware, error) {
	// TODO: Use the armored witness transparency logs as the source of firmware images.
	// For now, we'll just expect that repos are checked-out in adjacent directories,
	// and images have been built there.
	fw := &firmware{}

	var err error
	if fw.Recovery, err = os.ReadFile(*recoveryImagePath); err != nil {
		return nil, fmt.Errorf("failed to read recovery image from %q: %v", *recoveryImagePath, err)
	}
	if fw.BootLoader, err = os.ReadFile(*bootloaderPath); err != nil {
		return nil, fmt.Errorf("failed to read bootloader from %q: %v", *bootloaderPath, err)
	}
	if fw.TrustedApplet, err = os.ReadFile(*trustedAppletPath); err != nil {
		return nil, fmt.Errorf("failed to read trusted applet from %q: %v", *trustedAppletPath, err)
	}
	if fw.TrustedOS, err = os.ReadFile(*trustedOSPath); err != nil {
		return nil, fmt.Errorf("failed to read trusted OS from %q: %v", *trustedOSPath, err)
	}

	klog.Info("Loaded firmware artefacts:")
	klog.Infof("Recovery:      SHA256:%032x (%s)", sha256.Sum256(fw.Recovery), *recoveryImagePath)
	klog.Infof("Bootloader:    SHA256:%032x (%s)", sha256.Sum256(fw.BootLoader), *bootloaderPath)
	klog.Infof("TrustedApplet: SHA256:%032x (%s)", sha256.Sum256(fw.TrustedApplet), *trustedAppletPath)
	klog.Infof("TrustedOS:     SHA256:%032x (%s)", sha256.Sum256(fw.TrustedOS), *trustedOSPath)

	return fw, nil
}

func waitAndProvision(fw *firmware) error {
	// The device will initially be in HID mode (showing as "RecoveryMode" in the output to lsusb).
	// So we'll detect it as such:
	target, err := waitForHIDDevice()
	if err != nil {
		return err
	}

	_, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate ephemeral key: %v", err)
	}

	// Per-device prep:
	// TODO: sign bootloader and recovery images.
	// TODO: store signed bootloader and recovery images somewhere durable.

	// TODO: SDP boot recovery image on device.
	if err := target.BootIMX(fw.Recovery); err != nil {
		klog.Errorf("Failed to SDP boot recovery image on %v: %v", target.DeviceInfo.Path, err)
	}
	// TODO: figure out corresponding block device once it boots.
	// TODO: Write bootloader.
	// TODO: Write TrustedOS.
	// TODO: Write TrustedApplet.
	// TODO: Write proof bundle.

	// TODO: Verify fuses are unset.
	// TODO: Set fuses.

	// TODO: Reboot device.

	// TODO: Use HID to access witness public keys from device and store somewhere durable.

	return nil

}

// waitForHIDDevice waits for an unprovisioned armored witness device
// to appear on the USB bus.
func waitForHIDDevice() (*Target, error) {
	klog.Info("Waiting for device to be detected...")
	for {
		<-time.After(time.Second)
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
