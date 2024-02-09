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

package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/sumdb/note"
)

// createCmd represents the create command
var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new manifest file",
		Long:  `This command is used to create a new firmware manifest file.\n\nThe manifest describes important details about the firmware. `,
		Run:   create,
	}

	// knownFirmwareTypes is the set of possible values for the firmware_type flag.
	knownFirmwareTypes = map[string]struct{}{
		ftlog.ComponentApplet:   {},
		ftlog.ComponentBoot:     {},
		ftlog.ComponentOS:       {},
		ftlog.ComponentRecovery: {},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().String("git_tag", "", "The semantic version of the Trusted Applet release.")
	createCmd.Flags().String("git_commit_fingerprint", "", "Hex-encoded SHA-1 commit hash of the git repository when checked out at the specified git_tag.")
	createCmd.Flags().String("firmware_file", "", "Path of the firmware ELF file. ")
	createCmd.Flags().String("tamago_version", "", "The version of the Tamago (https://github.com/usbarmory/tamago) used to compile the Trusted Applet.")
	createCmd.Flags().String("output_file", "", "The file to write the manifest to. If this is not set, then only print the manifest to stdout.")
	createCmd.Flags().String("firmware_type", "", fmt.Sprintf("One of %v ", maps.Keys(knownFirmwareTypes)))
	createCmd.Flags().String("hab_target", "", "The devices the --hab_signature is targeting.")
	createCmd.Flags().String("hab_signature_file", "", "The HAB signature for the firmware file.")
	createCmd.Flags().Bool("raw", false, "If set, the command only emits the raw manifest JSON, it will not sign and encapsulate into a note")
	createCmd.Flags().String("private_key_file", "", "The file containing a Note-formatted signer string, used to sign the manifest")
}

func create(cmd *cobra.Command, args []string) {
	gitTag := requireFlagString(cmd.Flags(), "git_tag")
	gitCommitFingerprint := requireFlagString(cmd.Flags(), "git_commit_fingerprint")
	firmwareFile := requireFlagString(cmd.Flags(), "firmware_file")
	tamagoVersion := requireFlagString(cmd.Flags(), "tamago_version")
	firmwareType := requireFlagString(cmd.Flags(), "firmware_type")
	if _, ok := knownFirmwareTypes[firmwareType]; !ok {
		log.Fatalf("firmware_type must be one of %v", maps.Keys(knownFirmwareTypes))
	}
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		log.Fatal(err)
	}
	firmwareBytes, err := os.ReadFile(firmwareFile)
	if err != nil {
		log.Fatalf("Failed to read firmware_file %q: %v", firmwareFile, err)
	}
	digestBytes := sha256.Sum256(firmwareBytes)

	gitTagName, err := semver.NewVersion(gitTag)
	if err != nil {
		log.Fatalf("Failed to parse git_tag: %v", err)
	}
	tamagoVersionName, err := semver.NewVersion(tamagoVersion)
	if err != nil {
		log.Fatalf("Failed to parse tamago_version: %v", err)
	}
	r := ftlog.FirmwareRelease{
		Component:            firmwareType,
		GitTagName:           *gitTagName,
		GitCommitFingerprint: gitCommitFingerprint,
		FirmwareDigestSha256: digestBytes[:],
		TamagoVersion:        *tamagoVersionName,
	}
	if firmwareType == ftlog.ComponentBoot || firmwareType == ftlog.ComponentRecovery {
		habSigFile := requireFlagString(cmd.Flags(), "hab_signature_file")
		habSig, err := os.ReadFile(habSigFile)
		if err != nil {
			log.Fatalf("Failed to read HAB signature file %q: %v", habSigFile, err)
		}
		r.HAB = ftlog.HAB{
			Target:    requireFlagString(cmd.Flags(), "hab_target"),
			Signature: habSig,
		}
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// Note requires the msg ends in a newline.
	b = append(b, byte('\n'))

	if !raw {
		keyFile := requireFlagString(cmd.Flags(), "private_key_file")
		signer, err := os.ReadFile(keyFile)
		if err != nil {
			log.Fatalf("Failed to read private key file: %v", err)
		}
		b, err = sign(string(signer), b)
		if err != nil {
			log.Fatalf("Failed to sign manifest: %v", err)
		}
	}

	fmt.Print(string(b))

	outputFile, _ := cmd.Flags().GetString("output_file")
	if outputFile == "" {
		return
	}
	if err := os.WriteFile(outputFile, b, 0664); err != nil {
		log.Fatal(err)
	}
}

func requireFlagString(f *pflag.FlagSet, name string) string {
	v, err := f.GetString(name)
	if err != nil {
		log.Fatalf("Getting flag %v: %v", name, err)
	}
	if v == "" {
		log.Fatalf("Flag %v must be speficied", name)
	}
	return v
}

func sign(sec string, b []byte) ([]byte, error) {
	signer, err := note.NewSigner(sec)
	if err != nil {
		return nil, err
	}
	t := string(b)
	return note.Sign(&note.Note{Text: t}, signer)
}
