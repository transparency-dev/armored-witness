// Copyright 2023 Google LLC. All Rights Reserved.
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
package build

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/sumdb/note"
)

func NewReleaseImplicitMetadata(logV, osV1, osV2, appV, bootV, recoveryV string) (*ReleaseImplicitMetadata, error) {
	osReleaseVerifier1, err := note.NewVerifier(osV1)
	if err != nil {
		return nil, fmt.Errorf("failed to construct OS release verifier: %v", err)
	}
	osReleaseVerifier2, err := note.NewVerifier(osV2)
	if err != nil {
		return nil, fmt.Errorf("failed to construct OS release verifier: %v", err)
	}
	appletReleaseVerifier, err := note.NewVerifier(appV)
	if err != nil {
		return nil, fmt.Errorf("failed to construct applet release verifier: %v", err)
	}
	bootReleaseVerifier, err := note.NewVerifier(bootV)
	if err != nil {
		return nil, fmt.Errorf("failed to construct boot release verifier: %v", err)
	}
	recoveryReleaseVerifier, err := note.NewVerifier(recoveryV)
	if err != nil {
		return nil, fmt.Errorf("failed to construct recovery release verifier: %v", err)
	}
	allV := note.VerifierList(osReleaseVerifier1, osReleaseVerifier2, appletReleaseVerifier, bootReleaseVerifier, recoveryReleaseVerifier)

	dir, err := os.MkdirTemp("", "armored-witness-build-keys")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	logFile := filepath.Join(dir, "pubkey_log.pub")
	os1File := filepath.Join(dir, "pubkey_os1.pub")
	os2File := filepath.Join(dir, "pubkey_os2.pub")
	appletFile := filepath.Join(dir, "pubkey_applet.pub")
	if err := os.WriteFile(logFile, []byte(logV), 0644); err != nil {
		return nil, fmt.Errorf("failed to create public key file: %v", err)
	}
	if err := os.WriteFile(os1File, []byte(osV1), 0644); err != nil {
		return nil, fmt.Errorf("failed to create public key file: %v", err)
	}
	if err := os.WriteFile(os2File, []byte(osV2), 0644); err != nil {
		return nil, fmt.Errorf("failed to create public key file: %v", err)
	}
	if err := os.WriteFile(appletFile, []byte(appV), 0644); err != nil {
		return nil, fmt.Errorf("failed to create public key file: %v", err)
	}

	return &ReleaseImplicitMetadata{
		OSV1:      osReleaseVerifier1,
		OSV2:      osReleaseVerifier2,
		AppV:      appletReleaseVerifier,
		BootV:     bootReleaseVerifier,
		RecoveryV: recoveryReleaseVerifier,
		AllV:      allV,
		Envs: []string{
			fmt.Sprintf("LOG_PUBLIC_KEY=%s", logFile),
			fmt.Sprintf("APPLET_PUBLIC_KEY=%s", appletFile),
			fmt.Sprintf("OS_PUBLIC_KEY1=%s", os1File),
			fmt.Sprintf("OS_PUBLIC_KEY2=%s", os2File),
		},
		Cleanup: func() {
			os.RemoveAll(dir)
		},
	}, nil
}

// ReleaseImplicitMetadata stores all of the information needed to reproduce and
// verify releases. This is all of the data that is not passed in-band with the
// release (i.e. is not in the Makefile or code).
// In order to be maximally useful this exposes its state as env variables, which
// is how they are consumed. Some of these point at files, which need to be cleaned
// up after usage. This cleanup must be done by the owner of this object via the
// cleanup function.
type ReleaseImplicitMetadata struct {
	OSV1      note.Verifier
	OSV2      note.Verifier
	AppV      note.Verifier
	BootV     note.Verifier
	RecoveryV note.Verifier
	AllV      note.Verifiers
	Envs      []string
	Cleanup   func()
}
