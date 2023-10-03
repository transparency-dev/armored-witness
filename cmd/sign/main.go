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
// It is intended to be used to sign a manifest file for the Armored Witness
// firmware transparency log.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/transparency-dev/armored-witness/pkg/kmssigner"

	"cloud.google.com/go/kms/apiv1"
	"golang.org/x/mod/sumdb/note"
)

func main() {
	gcpProject := flag.String("project_name", "",
		"The GCP project name where the signing key lives.")
	keyRing := flag.String("key_ring", "",
		"Key ring of the signing key. See https://cloud.google.com/kms/docs/resource-hierarchy#key_rings.")
	keyName := flag.String("key_name", "",
		"Name of the signing key in the key ring.")
	keyVersion := flag.Uint("key_version", 0,
		"Version of the signing key. See https://cloud.google.com/kms/docs/resource-hierarchy#key_versions")
	keyLocation := flag.String("key_location", "",
		"Location (GCP region) of the signing key.")
	manifestFile := flag.String("manifest_file", "",
		"The file containing the content to sign.")
	outputFile := flag.String("output_file", "",
		"The file to write the note to.")

	flag.Parse()

	if *gcpProject == "" {
		log.Fatal("project_name is required.")
	}
	if *keyRing == "" {
		log.Fatal("key_ring is required.")
	}
	if *keyName == "" {
		log.Fatal("key_name is required.")
	}
	if *keyVersion == 0 {
		log.Fatal("key_version must be > 0.")
	}
	if *keyLocation == "" {
		log.Fatal("key_location is required.")
	}
	if *manifestFile == "" {
		log.Fatal("manifest_file is required.")
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

	kmsKeyVersionResourceName := fmt.Sprintf(kmssigner.KeyVersionNameFormat, *gcpProject, *keyLocation,
		*keyRing, *keyName, *keyVersion)
	signer, err := kmssigner.New(ctx, kmClient, kmsKeyVersionResourceName)
	if err != nil {
		log.Fatalf("failed to create signer: %v", err)
	}

	// Sign manifestFile as note.
	manifestBytes, err := os.ReadFile(*manifestFile)
	if err != nil {
		log.Fatalf("failed to read manifest_file %q: %v", *manifestFile, err)
	}
	msg, err := note.Sign(&note.Note{Text: string(manifestBytes)}, signer)
	if err != nil {
		log.Fatalf("failed to sign note text from %q: %v", *manifestFile, err)
	}

	// Write output file.
	if err := os.WriteFile(*outputFile, msg, 0664); err != nil {
		log.Fatalf("failed to write outputFile %q: %v", *outputFile, err)
	}
}
