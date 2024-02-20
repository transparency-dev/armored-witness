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

//go:build !tamago
// +build !tamago

package device

import (
	"fmt"

	flynn_hid "github.com/flynn/hid"
	"github.com/flynn/u2f/u2fhid"
	"google.golang.org/protobuf/proto"

	"github.com/transparency-dev/armored-witness-os/api"
)

// DetectU2F returns the first U2F device found which matches
// the armored witness vendor and product IDs.
func DetectU2F() (string, *u2fhid.Device, error) {
	devices, err := flynn_hid.Devices()
	if err != nil {
		return "", nil, err
	}

	for _, d := range devices {
		if d.UsagePage == api.HIDUsagePage &&
			d.VendorID == api.VendorID &&
			d.ProductID == api.ProductID {

			dev, err := u2fhid.Open(d)
			return d.Path, dev, err
		}
	}

	return "", nil, nil
}

// WitnessStatus issues the Status command to the armored witness via HID and returns the result.
func WitnessStatus(dev *u2fhid.Device) (*api.Status, error) {
	res, err := dev.Command(api.U2FHID_ARMORY_INF, nil)
	if err != nil {
		return nil, err
	}

	s := &api.Status{}
	if err := proto.Unmarshal(res, s); err != nil {
		return nil, err
	}

	return s, nil
}

// ActivateHAB issues the HAB command to the armored witness via HID.
func ActivateHAB(dev *u2fhid.Device) error {
	buf, err := dev.Command(api.U2FHID_ARMORY_HAB, nil)
	if err != nil {
		return err
	}

	res := &api.Response{}
	if err := proto.Unmarshal(buf, res); err != nil {
		return err
	}
	if res.Error != api.ErrorCode_NONE {
		return fmt.Errorf("%v: %s", res.Error, res.Payload)
	}
	return nil
}
