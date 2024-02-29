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

// Package device contains functions for dealing with the armored witness
// hardware device.
package device

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

// BootIntoRecovery attempts to boot an armored witness device in SDP (Serial Boot) mode
// with the recovery firmware image, and then watches for a matching new block device
// to be presented to the host.
//
// If no armored witness device is present, this function will wait until either a device
// is plugged in/rebooted into SDP mode by the user, or the context becomes done.
//
// Returns the HID device and detected block device path, or an error.
func BootIntoRecovery(ctx context.Context, recoveryFirmware []byte, blockDeviceGlob string) (*Target, string, error) {
	target, err := waitForHIDDevice(ctx)
	if err != nil {
		return nil, "", err
	}

	// SDP boot recovery image on device.
	// Booting the recovery image causes the device re-appear as a USB Mass Storage device.
	// So we'll wait for that to happen, and figure out which /dev/ entry corresponds to it.
	bDev, err := waitForBlockDevice(ctx, blockDeviceGlob, func() error {
		if err := target.BootIMX(recoveryFirmware); err != nil {
			return fmt.Errorf("failed to SDP boot recovery image on %v: %v", target.DeviceInfo.Path, err)
		}
		klog.Info("Witness device booting recovery image")
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to detect block device: %v", err)

	}
	return target, bDev, nil
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
	if err := watcher.Add("/dev/disk/by-id"); err != nil {
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
			matched, err := filepath.Match(glob, e.Name)
			if err != nil {
				klog.Exitf("error testing filename %q against glob %q: %v", e.Name, glob, err)
			}
			if matched && e.Has(fsnotify.Create) {
				// At least on linux, it takes a while for the device to become usable
				// TODO: can we detect when it's usable directly?
				klog.Info("Waiting for block device to settle...")
				time.Sleep(5 * time.Second)
				return e.Name, nil
			}
		}
	}
}
