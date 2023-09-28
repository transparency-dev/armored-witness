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
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/mod/sumdb/note"
)

// verifysigCmd represents the verifysig command
var verifysigCmd = &cobra.Command{
	Use:   "verifysig",
	Short: "Verify a manifest has a valid signature from a given key",
	Long:  `This command verifies that a manifest is signed by a specific public key.\n\nTo check that valid signatures from multiple keys are present, re-run this command with each of the public keys in turn`,
	Run:   verify,
}

func init() {
	rootCmd.AddCommand(verifysigCmd)

	verifysigCmd.Flags().String("input_file", "", "The file to read the signed manifest from. If this is not set, then read the manifest from stdin.")
	verifysigCmd.Flags().String("public_key", "", "Note-formatted verifier string, used to verify the signature on the manifest")
}

func verify(cmd *cobra.Command, args []string) {
	pubK := requireFlagString(cmd.Flags(), "public_key")
	verifier, err := note.NewVerifier(pubK)
	if err != nil {
		log.Fatalf("Invalid public key: %v", err)
	}

	msg := []byte{}
	input, _ := cmd.Flags().GetString("input_file")
	if input == "" {
		msg, err = io.ReadAll(os.Stdin)
	} else {
		input = "<stdin>"
		msg, err = os.ReadFile(input)
	}
	if err != nil {
		log.Fatalf("Failed to read manifest from %v: %v", input, err)
	}
	if _, err = note.Open(msg, note.VerifierList(verifier)); err != nil {
		log.Fatalf("Failed to open manifest: %v", err)
	}
	log.Println("Manifest signature verified ok")
}
