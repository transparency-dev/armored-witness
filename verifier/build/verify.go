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
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"k8s.io/klog/v2"
)

const (
	gitOwner      = "transparency-dev"
	gitRepoApplet = "armored-witness-applet"
	gitRepoOS     = "armored-witness-os"
	gitRepoBoot   = "armored-witness-boot"
)

// NewReproducibleBuildVerifier returns a ReproducibleBuildVerifier that will delete
// any temporary git repositories after use if cleanup is true, or leave them around
// for further investigation if false.
func NewReproducibleBuildVerifier(cleanup bool, tamago Tamago, sigs releaseImplicitMetadata) (*ReproducibleBuildVerifier, error) {
	return &ReproducibleBuildVerifier{
		cleanup: cleanup,
		tamago:  tamago,
		sigs:    sigs,
	}, nil
}

// ReproducibleBuildVerifier checks out the source code referenced by a manifest and
// determines whether it can reproduce the final build artifacts.
type ReproducibleBuildVerifier struct {
	cleanup bool
	tamago  Tamago
	sigs    releaseImplicitMetadata
}

// VerifyManifest attempts to reproduce the FirmwareRelease at index `i` in the log by
// checking out the code and running the make file.
func (v *ReproducibleBuildVerifier) VerifyManifest(ctx context.Context, i uint64, r ftlog.FirmwareRelease) error {
	klog.V(1).Infof("VerifyManifest %d: %s@%s", i, r.Component, r.GitTagName)
	var cv componentVerifier
	switch r.Component {
	case ftlog.ComponentApplet:
		cv = appletVerifier{}
	case ftlog.ComponentOS:
		cv = osVerifier{}
	case ftlog.ComponentBoot:
		cv = bootVerifier{}
	default:
		return fmt.Errorf("Unsupported component: %q", r.Component)
	}

	// Download, install, and then set up tamago environment to match manifest
	if err := v.tamago.Switch(r.TamagoVersion); err != nil {
		return fmt.Errorf("failed to switch tamago version: %v", err)
	}

	// Create temporary directory that will be cleaned up after this method returns
	dir, err := os.MkdirTemp("", "armored-witness-build-verify")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Cleanup will occur:
	//  - if v.cleanup; OR
	//  - if !v.cleanup AND build succeeds
	// This means that disabling v.cleanup will only leave failed builds around for forensics.
	cleanup := v.cleanup
	defer func() {
		if cleanup {
			os.RemoveAll(dir)
		} else {
			klog.Infof("🔎 Evidence of failed build: %s (%d: %s@%s)", dir, i, r.Component, r.GitCommitFingerprint)
		}
	}()

	klog.V(1).Infof("Cloning repo into %q", dir)

	// Clone the repository at the release tag
	// TODO(mhutchinson): this should check out the GitTagName but we don't tag
	// all releases in CI.
	// 	cmd := exec.Command("/usr/bin/git", "clone", fmt.Sprintf("https://github.com/%s/%s", gitOwner, repo), "-b", r.GitTagName)
	repo := cv.repo()
	cmd := exec.Command("/usr/bin/git", "clone", cv.repo())
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone: %v (%s)", err, out)
	}

	repoRoot := filepath.Join(dir, repo[strings.LastIndex(repo, "/"):])

	cmd = exec.Command("/usr/bin/git", "reset", "--hard", r.GitCommitFingerprint)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset to commit: %v (%s)", err, out)
	}

	// Confirm that the git revision matches the manifest
	// This has been left in with the expectation that we check out by TagName above
	// at some point. If we don't do that then this is pretty redundant.
	cmd = exec.Command("/usr/bin/git", "rev-parse", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get HEAD revision: %v (%s)", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), r.GitCommitFingerprint; got != want {
		return fmt.Errorf("expected revision %q but got %q for tag %q", want, got, r.GitTagName)
	}

	// Make the elf file
	cmd = cv.makeCommand()
	cmd.Dir = repoRoot
	cmd.Env = append(cmd.Env, v.tamago.Envs(r.TamagoVersion)...)
	cmd.Env = append(cmd.Env, v.sigs.envs...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SEMVER_TAG=%s", r.GitTagName))
	cmd.Env = append(cmd.Env, r.BuildEnvs...)
	klog.V(1).Infof("Running %q in %s", cmd.String(), repoRoot)
	if klog.V(2).Enabled() {
		for _, e := range cmd.Env {
			klog.V(2).Infof("  %s", e)
		}
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to make: %v (%s)", err, out)
	} else if klog.V(2).Enabled() {
		klog.V(2).Info(string(out))
	}

	// Hash the firmware artifact.
	data, err := os.ReadFile(filepath.Join(repoRoot, cv.binFile()))
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", cv.binFile(), err)
	}
	if got, want := sha256.Sum256(data), r.FirmwareDigestSha256; !bytes.Equal(got[:], want) {
		// TODO: report this in a more visible way than an error in the log.
		klog.Errorf("Leaf index %d: ❌ failed to reproduce build %s@%s (%s) => (got %x, wanted %x)", i, r.Component, r.GitTagName, r.GitCommitFingerprint, got, want)
		return nil
	}

	klog.Infof("Leaf index %d: ✅ reproduced build %s@%s (%s) => %x", i, r.Component, r.GitTagName, r.GitCommitFingerprint, r.FirmwareDigestSha256)
	cleanup = true
	return nil
}

type componentVerifier interface {
	repo() string
	makeCommand() *exec.Cmd
	binFile() string
}
type appletVerifier struct {
}

func (v appletVerifier) repo() string {
	return fmt.Sprintf("https://github.com/%s/%s", gitOwner, gitRepoApplet)
}

func (v appletVerifier) makeCommand() *exec.Cmd {
	return exec.Command("/usr/bin/make", "trusted_applet_nosign")
}

func (v appletVerifier) binFile() string {
	return "bin/trusted_applet.elf"
}

type osVerifier struct {
}

func (v osVerifier) repo() string {
	return fmt.Sprintf("https://github.com/%s/%s", gitOwner, gitRepoOS)
}

func (v osVerifier) makeCommand() *exec.Cmd {
	return exec.Command("/usr/bin/make", "trusted_os_release")
}

func (v osVerifier) binFile() string {
	return "bin/trusted_os.elf"
}

type bootVerifier struct {
}

func (v bootVerifier) repo() string {
	return fmt.Sprintf("https://github.com/%s/%s", gitOwner, gitRepoBoot)
}

func (v bootVerifier) makeCommand() *exec.Cmd {
	return exec.Command("/usr/bin/make", "imx")
}

func (v bootVerifier) binFile() string {
	return "armored-witness-boot.imx"
}
