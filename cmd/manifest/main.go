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
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-semver/semver"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
)

func main() {
	gitTag := flag.String("git_tag", "",
		"The semantic version of the Trusted Applet release.")
	gitCommitFingerprint := flag.String("git_commit_fingerprint", "",
		"Hex-encoded SHA-1 commit hash of the git repository when checked out at the specified git_tag.")
	firmwareFile := flag.String("firmware_file", "",
		"Path of the firmware ELF file. ")
	tamagoVersion := flag.String("tamago_version", "",
		"The version of the Tamago (https://github.com/usbarmory/tamago) used to compile the Trusted Applet.")

	flag.Parse()

	if *gitTag == "" {
		log.Fatal("git_tag is required.")
	}
	if *gitCommitFingerprint == "" {
		log.Fatal("git_commit_fingerprint is required.")
	}
	if *firmwareFile == "" {
		log.Fatal("firmware_file is required.")
	}
	if *tamagoVersion == "" {
		log.Fatal("tamago_version is required.")
	}

	firmwareBytes, err := os.ReadFile(*firmwareFile)
	if err != nil {
		log.Fatalf("Failed to read firmware_file %q: %v", *firmwareFile, err)
	}
	digestBytes := sha256.Sum256(firmwareBytes)

	r := ftlog.FirmwareRelease{
		Component:            ftlog.ComponentApplet,
		GitTagName:           *semver.New(*gitTag),
		GitCommitFingerprint: *gitCommitFingerprint,
		FirmwareDigestSha256: digestBytes[:],
		TamagoVersion:        *semver.New(*tamagoVersion),
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
