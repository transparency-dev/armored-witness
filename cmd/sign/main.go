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
// The sign tool signs an input file in the
// [note](https://pkg.go.dev/golang.org/x/mod/sumdb/note) format with a key
// from Google Cloud Platform's
// [Key Management Service](https://cloud.google.com/kms/docs).
//
// It is intended to be used to sign/cosign a manifest file for the Armored
// Witness firmware transparency log.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/transparency-dev/armored-witness/pkg/kmssigner"

	kms "cloud.google.com/go/kms/apiv1"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/sumdb/note"
)

const usageString = `
This program is used to sign manifests using the note format.

It can be used to create an initial signed manifest directly from a JSON file,
or, it can be used to counter-sign an existing signed manifest.

To create a signed manifest, the --manifest_file flag must be supplied:
$ sign --manifest_file=<path to manifest> --output=<path to output>

To counter sign a manifest, the --note_file and --note_verifier flags must
be supplied:
$ sign --note_file=<path to previously signed manifest> --note_verifier=<verifier string for previous signature>
`

// keyInfo represents a KMS key and its corresponding public verifier.
type keyInfo struct {
	kmsKeyName    string
	kmsKeyVersion uint
	noteVerifier  string
}

var (
	// signingCfg holds the KMS and note parameters for the various artefacts keyed by release environment.
	signingCfg = map[string]struct {
		kmsKeyRing string
		kmsRegion  string
		keys       map[string]keyInfo
	}{
		"ci": {
			kmsKeyRing: "firmware-release-ci",
			kmsRegion:  "global",
			keys: map[string]keyInfo{
				"ftlog": {
					kmsKeyName:    "ft-log-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-ftlog-ci+f5479c1e+AR6gW0mycDtL17iM2uvQUThJsoiuSRirstEj9a5AdCCu",
				},
				"applet": {
					kmsKeyName:    "trusted-applet-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3",
				},
				"boot": {
					kmsKeyName:    "bootloader-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-boot-ci+9f62b6ac+AbnipFmpRltfRiS9JCxLUcAZsbeH4noBOJXbVD3H5Eg4",
				},
				"os1": {
					kmsKeyName:    "trusted-os-1-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ",
				},
				"os2": {
					kmsKeyName:    "trusted-os-2-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh",
				},
				"recovery": {
					kmsKeyName:    "recovery-ci",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-recovery-ci+cc699423+AarlJMSl0rbTMf31B5o9bqc6PHorwvF1GbwyJRXArbfg",
				},
			},
		},
		"prod": {
			kmsKeyRing: "firmware-release-prod",
			kmsRegion:  "global",
			keys: map[string]keyInfo{
				"ftlog": {
					kmsKeyName:    "ft-log-prod",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-ftlog-prod+72b0da75+Aa3qdhefd2cc/98jV3blslJT2L+iFR8WKHeGcgFmyjnt",
				},
				"applet": {
					kmsKeyName:    "trusted-applet-prod",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-applet-prod+d45f2a0d+AZSnFa8GxH+jHV6ahELk6peqVObbPKrYAdYyMjrzNF35",
				},
				"boot": {
					kmsKeyName:    "bootloader-prod",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-boot-prod+2fa9168e+AR+KIx++GIlMBICxLkf4ZUK5RDlvJuiYUboqX5//RmUm",
				},
				"os": {
					kmsKeyName:    "trusted-os-prod",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-os-prod+c31218b7+AV7mmRamQp6VC9CutzSXzqtNhYNyNmQQRcLX07F6qlC1",
				},
				"recovery": {
					kmsKeyName:    "recovery-prod",
					kmsKeyVersion: 1,
					noteVerifier:  "transparency.dev-aw-recovery-prod+f3710baa+ATu+HMUuO8ZsgaNwP97XMcb/+Ve8W1u1KdFQHNzOyLxx",
				},
			},
		},
	}
)

func main() {
	gcpProject := flag.String("project_name", "",
		"The GCP project name where the signing key lives.")
	release := flag.String("release", "", fmt.Sprintf("Release type, one of: %v", maps.Keys(signingCfg)))
	artefact := flag.String("artefact", "", "Type of artefact being signed, one of: "+artefactTypes())
	manifestFile := flag.String("manifest_file", "",
		"The file containing the content to sign.")
	noteFile := flag.String("note_file", "", "The file containing a note to cosign.")
	noteVerifier := flag.String("note_verifier", "", "If cosigning an existing note, this verifier string is used to verify the note before countersigning.")
	outputFile := flag.String("output_file", "",
		"The file to write the note to.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), usageString+"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *gcpProject == "" {
		log.Fatal("project_name is required.")
	}
	rel, ok := signingCfg[*release]
	if !ok {
		log.Fatalf("release is required, and must be one of %v", maps.Keys(signingCfg))
	}
	keyInfo, ok := rel.keys[*artefact]
	if !ok {
		log.Fatalf("artefact is required and must be one of %v", artefactTypes())
	}
	if len(*manifestFile) == len(*noteFile) {
		log.Fatalf("either manifest_file or note_file must be provided")
	}
	if *outputFile == "" {
		log.Fatal("output_file is required.")
	}

	ctx := context.Background()
	kmClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatalf("failed to create KeyManagementClient: %v", err)
	}
	defer kmClient.Close()

	verifier, err := note.NewVerifier(keyInfo.noteVerifier)
	if err != nil {
		log.Fatalf("invalid note verifier for %s/%s: %v", *release, *artefact, err)
	}
	kmsKeyVersionResourceName := fmt.Sprintf(kmssigner.KeyVersionNameFormat, *gcpProject, rel.kmsRegion,
		rel.kmsKeyRing, keyInfo.kmsKeyName, keyInfo.kmsKeyVersion)
	signer, err := kmssigner.New(ctx, kmClient, kmsKeyVersionResourceName, verifier.Name())
	if err != nil {
		log.Fatalf("failed to create signer for %s/%s: %v", *release, *artefact, err)
	}

	var n *note.Note
	switch {
	case *manifestFile != "":
		// Sign manifestFile as note.
		manifestBytes, err := os.ReadFile(*manifestFile)
		if err != nil {
			log.Fatalf("failed to read manifest_file %q: %v", *manifestFile, err)
		}
		n = &note.Note{Text: string(manifestBytes)}
	case *noteFile != "":
		nRaw, err := os.ReadFile(*noteFile)
		if err != nil {
			log.Fatalf("failed to read note_file %q: %v", *noteFile, err)
		}
		nV, err := note.NewVerifier(*noteVerifier)
		if err != nil {
			log.Fatalf("note_verifier: %v", err)
		}
		n, err = note.Open(nRaw, note.VerifierList(nV))
		if err != nil {
			log.Fatalf("failed to open note from %q: %v", *noteFile, err)
		}
	}

	msg, err := note.Sign(n, signer)
	if err != nil {
		log.Fatalf("failed to sign note text from %q: %v", *manifestFile, err)
	}

	// Verify signature was made by expected key
	if _, err := note.Open(msg, note.VerifierList(verifier)); err != nil {
		log.Fatalf("failed to verify signature for %s/%s: %v", *release, *artefact, err)
	}

	// Write output file.
	if err := os.WriteFile(*outputFile, msg, 0664); err != nil {
		log.Fatalf("failed to write outputFile %q: %v", *outputFile, err)
	}
}

func artefactTypes() string {
	r := ""
	for k, v := range signingCfg {
		r += fmt.Sprintf("%s: %v", k, maps.Keys(v.keys))
	}
	return r
}
