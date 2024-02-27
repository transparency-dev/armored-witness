// Copyright 2024 The Armored Witness authors. All Rights Reserved.
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
// The printverifier command prints a note compatible verifier string for
// a public key stored in GCP.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"golang.org/x/mod/sumdb/note"
)

var (
	keyResource = flag.String("key", "", "GCP Resource ID for the public key to use")
	keyName     = flag.String("name", "", "Template for created note verifier's name, use %%d to include keyRevision")
)

func main() {
	flag.Parse()

	v, err := fromGCP(context.Background(), *keyResource, *keyName)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("%s\n", v)
}

func fromGCP(ctx context.Context, f string, n string) (string, error) {
	m := regexp.MustCompile("projects/.*/locations/.*/keyRings/.*/cryptoKeys/.*/cryptoKeyVersions/(.*)")
	if !m.MatchString(f) {
		return "", fmt.Errorf("invalid GCP key name")
	}
	c, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create KeyManagementClient: %v", err)
	}
	go func() {
		<-ctx.Done()
		c.Close()
	}()

	resp, err := c.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: f})
	if err != nil {
		return "", fmt.Errorf("GetPublicKey: %v", err)
	}
	der, _ := pem.Decode([]byte(resp.GetPem()))
	pk, err := x509.ParsePKIXPublicKey(der.Bytes)
	if err != nil {
		return "", fmt.Errorf("ParsePKIXPublicKey: %v", err)
	}
	edk, ok := pk.(ed25519.PublicKey)
	if !ok {
		return "", fmt.Errorf("Oh noes, got a %T but want an Ed25519 key", pk)
	}

	if strings.Contains(n, "%") {
		var version int
		if vs := m.FindStringSubmatch(f); len(vs) > 1 {
			if v, err := strconv.Atoi(vs[1]); err != nil {
				return "", fmt.Errorf("couldn't parse keyVersion: %v", err)
			} else {
				version = v
			}
		}
		n = fmt.Sprintf(n, version)
	}
	vkey, err := note.NewEd25519VerifierKey(n, edk)
	if err != nil {
		return "", fmt.Errorf("NewEd25519VerifierKey: %s: %v", n, err)
	}
	if _, err := note.NewVerifier(vkey); err != nil {
		return "", fmt.Errorf("NewVerifier: %v", err)
	}
	return vkey, nil
}
