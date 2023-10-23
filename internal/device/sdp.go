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

// Large portions of this file come from
// https://github.com/usbarmory/armory-boot/blob/master/cmd/armory-boot-usb/armory-boot-usb.go

package device

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/usbarmory/armory-boot/sdp"
	"k8s.io/klog/v2"

	"github.com/usbarmory/hid"
)

const (
	// USB vendor ID for all supported devices
	FreescaleVendorID = 0x15a2

	// On-Chip RAM (OCRAM/iRAM) address for payload staging
	iramOffset = 0x00910000

	// SDP HID report IDs
	// (p327, 8.9.3.1 SDP commands, IMX6ULLRM).
	H2D_COMMAND       = 1 // Command  - Host to Device
	H2D_DATA          = 2 // Data     - Host to Device
	D2H_RESPONSE      = 3 // Response - Device to Host
	D2H_RESPONSE_LAST = 4 // Response - Device to Host
)

var (
	// These are the only supported devices for the armored witness.
	supportedDevices = map[uint16]string{
		0x007d: "Freescale SemiConductor Inc  SE Blank 6UL",
		0x0080: "Freescale SemiConductor Inc  SE Blank 6ULL",
	}

	Timeout = 10 * time.Second
)

// Target represents a valid device to be SDP booted
type Target struct {
	sync.Mutex

	// DeviceInfo allows callers to inspect details of the device (e.g. to differentiate
	// between multiple units connected at once).
	DeviceInfo hid.DeviceInfo

	// dev is the opened device, or nil
	dev hid.Device
}

// DetectHID returns a list of compatible devices in SDP mode.
func DetectHID() ([]*Target, error) {
	devices, err := hid.Devices()
	if err != nil {
		return nil, err
	}

	var ret []*Target
	for _, d := range devices {
		d := d
		if d.VendorID != FreescaleVendorID {
			continue
		}

		if product, ok := supportedDevices[d.ProductID]; ok {
			klog.Infof("found device %04x:%04x %s", d.VendorID, d.ProductID, product)
		} else {
			continue
		}

		ret = append(ret, &Target{DeviceInfo: *d})
	}

	return ret, nil
}

// BootIMX attempts to use SDP to send an IMX image to the target, and boot it.
func (t *Target) BootIMX(imx []byte) error {
	t.Lock()
	var err error
	if t.dev, err = t.DeviceInfo.Open(); err != nil {
		return fmt.Errorf("failed to open device %q: %v", t.DeviceInfo.Path, err)
	}
	defer func() {
		if t.dev != nil {
			t.dev.Close()
		}
		t.dev = nil
		t.Unlock()
	}()

	klog.Infof("Attempting to SDP boot device %s", t.DeviceInfo.Path)

	ivt, err := sdp.ParseIVT(imx)
	if err != nil {
		return fmt.Errorf("failed to parse IVT: %v", err)
	}

	dcd, err := sdp.ParseDCD(imx, ivt)
	if err != nil {
		return fmt.Errorf("failed to parse DCD: %v", err)
	}

	klog.Infof("Loading DCD at %#08x (%d bytes)", iramOffset, len(dcd))
	if err = t.dcdWrite(dcd, iramOffset); err != nil {
		return fmt.Errorf("failed to write DCD: %v", err)
	}

	klog.Infof("Loading imx to %#08x (%d bytes)", ivt.Self, len(imx))
	if err = t.fileWrite(imx, ivt.Self); err != nil {
		return fmt.Errorf("failed to write IMX file: %v", err)
	}

	klog.Infof("Sending jump address to %#08x", ivt.Self)
	if err = t.jumpAddress(ivt.Self); err != nil {
		return fmt.Errorf("failed to set jump address: %v", err)
	}

	klog.Infof("Serial download on %s complete", t.DeviceInfo.Path)
	return nil
}

func (t *Target) sendHIDReport(reqID int, buf []byte, resID int) (res []byte, err error) {
	if err := t.dev.Write(append([]byte{byte(reqID)}, buf...)); err != nil {
		return nil, fmt.Errorf("failed to send HID report to device (%v): %v", t.DeviceInfo.Path, err)
	}
	if resID < 0 {
		return nil, nil
	}

	for {
		select {
		case res, ok := <-t.dev.ReadCh():
			if !ok {
				return nil, errors.New("error reading response")
			}

			if len(res) > 0 && res[0] == byte(resID) {
				return res, nil
			}
		case <-time.After(Timeout):
			return nil, errors.New("command timeout")
		}
	}
}

func (t *Target) dcdWrite(dcd []byte, addr uint32) error {
	r1, r2 := sdp.BuildDCDWriteReport(dcd, addr)

	if _, err := t.sendHIDReport(H2D_COMMAND, r1, -1); err != nil {
		return fmt.Errorf("failed to send first DCD write report: %v", err)
	}

	if _, err := t.sendHIDReport(H2D_DATA, r2, D2H_RESPONSE_LAST); err != nil {
		return fmt.Errorf("failed to send second DCD write report: %v", err)
	}

	return nil
}

func (t *Target) fileWrite(imx []byte, addr uint32) error {
	r1, r2 := sdp.BuildFileWriteReport(imx, addr)

	if _, err := t.sendHIDReport(H2D_COMMAND, r1, -1); err != nil {
		return fmt.Errorf("failed to send FileWriteReport r1: %v", err)
	}

	// Don't wait for report responses until we've sent the final block.
	resID := -1

	for i, r := range r2 {
		if i == len(r2)-1 {
			// We're now sending the final chunk of the imx, so wait for
			// report ID 4 - this report indicates completion of the
			// FileWrite request.
			resID = D2H_RESPONSE_LAST
		}
	send:
		_, err := t.sendHIDReport(H2D_DATA, r, resID)
		if err != nil && runtime.GOOS == "darwin" && err.Error() == "hid: general error" {
			// On macOS access contention with the OS causes
			// errors, as a workaround we retry from the transfer
			// that got caught up.
			select {
			case <-time.After(Timeout):
				return err
			default:
				off := uint32(i) * 1024
				r1 := &sdp.SDP{
					CommandType: sdp.WriteFile,
					Address:     addr + off,
					DataCount:   uint32(len(imx)) - off,
				}

				if _, err = t.sendHIDReport(H2D_COMMAND, r1.Bytes(), -1); err != nil {
					return fmt.Errorf("(retry) failed to send FileWriteReport r1 file bytes at 0x%x: %v", r1.Address, err)
				}

				goto send
			}
		}

		if err != nil {
			return fmt.Errorf("failed to send FileWriteReport r2[%d]: %v", i, err)
		}
	}

	return nil
}

func (t *Target) jumpAddress(addr uint32) error {
	r1 := sdp.BuildJumpAddressReport(addr)
	if _, err := t.sendHIDReport(H2D_COMMAND, r1, -1); err != nil {
		return fmt.Errorf("failed to send JumpAddressReport: %v", err)
	}

	return nil
}
