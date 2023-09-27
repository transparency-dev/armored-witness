// Copyright 2023 The Armored Witness authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// The manifest tool formats input data into the Statement of the Armored
// Witness firmware transparency log.
package main

import (
	"github.com/transparency-dev/armored-witness/cmd/manifest/cmd"
)

// knownFirmwareTypes is the set of possible values for the firmware_type flag.
var knownFirmwareTypes = map[string]struct{}{
	ftlog.ComponentApplet:   {},
	ftlog.ComponentBoot:     {},
	ftlog.ComponentOS:       {},
	ftlog.ComponentRecovery: {},
}

func main() {
	cmd.Execute()
}
