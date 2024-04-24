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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"golang.org/x/mod/sumdb/note"
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
func NewReproducibleBuildVerifier(cleanup bool, tamago Tamago, metadata *ReleaseImplicitMetadata) (*ReproducibleBuildVerifier, error) {
	return &ReproducibleBuildVerifier{
		cleanup:  cleanup,
		tamago:   tamago,
		metadata: metadata,
	}, nil
}

// ReproducibleBuildVerifier checks out the source code referenced by a manifest and
// determines whether it can reproduce the final build artifacts.
type ReproducibleBuildVerifier struct {
	cleanup  bool
	tamago   Tamago
	metadata *ReleaseImplicitMetadata
}

// Verify checks everything that can be checked about a manifest in isolation:
//  1. That it is a valid note signed by the correct release signer
//  2. That this note contains a valid manifest file
//  3. That the binary committed to in the manifest file can be reproducibly built
//
// Returns true if the build was successfully reproduced, false otherwise, or an error if the build process itself failed.
func (v *ReproducibleBuildVerifier) Verify(ctx context.Context, i uint64, manifest []byte) (bool, error) {
	releaseNote, err := note.Open(manifest, v.metadata.AllV)
	if err != nil {
		if e, ok := err.(*note.UnverifiedNoteError); ok && len(e.Note.UnverifiedSigs) > 0 {
			return false, fmt.Errorf("unknown signer %q for leaf at index %d: %v", e.Note.UnverifiedSigs[0].Name, i, err)
		}
		return false, fmt.Errorf("failed to open leaf note at index %d: %v", i, err)
	}

	var release ftlog.FirmwareRelease
	if err := json.Unmarshal([]byte(releaseNote.Text), &release); err != nil {
		return false, fmt.Errorf("failed to unmarshal release at index %d: %w", i, err)
	}

	switch release.Component {
	case ftlog.ComponentApplet:
		if err := assertSigners(releaseNote, v.metadata.AppV); err != nil {
			return false, fmt.Errorf("applet sig verification failed: %v", err)
		}
	case ftlog.ComponentOS:
		if err := assertSigners(releaseNote, v.metadata.OSV1, v.metadata.OSV2); err != nil {
			return false, fmt.Errorf("os sig verification failed: %v", err)
		}
	case ftlog.ComponentBoot:
		if err := assertSigners(releaseNote, v.metadata.BootV); err != nil {
			return false, fmt.Errorf("boot sig verification failed: %v", err)
		}
	case ftlog.ComponentRecovery:
		if err := assertSigners(releaseNote, v.metadata.RecoveryV); err != nil {
			return false, fmt.Errorf("recovery sig verification failed: %v", err)
		}
	default:
		return false, fmt.Errorf("Unsupported component: %q", release.Component)
	}

	klog.V(1).Infof("Leaf index %d: verifying manifest: %s@%s (%s)", i, release.Component, release.Git.TagName, release.Git.CommitFingerprint)
	return v.verifyManifest(ctx, i, release)
}

// verifyManifest attempts to reproduce the FirmwareRelease at index `i` in the log by
// checking out the code and running the make file.
//
// Returns true if the build was successfully reproduced, false otherwise, or an error if the build process itself failed.
func (v *ReproducibleBuildVerifier) verifyManifest(ctx context.Context, i uint64, r ftlog.FirmwareRelease) (bool, error) {
	klog.V(1).Infof("verifyManifest %d: %s@%s", i, r.Component, r.Git.TagName)
	var cv componentVerifier
	switch r.Component {
	case ftlog.ComponentApplet:
		cv = appletVerifier{}
	case ftlog.ComponentOS:
		cv = osVerifier{}
	case ftlog.ComponentBoot:
		cv = bootVerifier{}
	case ftlog.ComponentRecovery:
		cv = recoveryVerifier{}
	default:
		return false, fmt.Errorf("Unsupported component: %q", r.Component)
	}

	// Download, install, and then set up tamago environment to match manifest
	if err := v.tamago.Switch(r.Build.TamagoVersion); err != nil {
		return false, fmt.Errorf("failed to switch tamago version: %v", err)
	}

	// Create temporary directory that will be cleaned up after this method returns
	dir, err := os.MkdirTemp("", "armored-witness-build-verify")
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %v", err)
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
			klog.Infof("🔎 Evidence of failed build: %s (%d: %s@%s)", dir, i, r.Component, r.Git.CommitFingerprint)
		}
	}()

	klog.V(1).Infof("Cloning repo into %q", dir)

	// Clone the repository at the release tag
	// TODO(mhutchinson): this should check out the GitTagName but we don't tag
	// all releases in CI.
	// 	cmd := exec.Command("/usr/bin/git", "clone", fmt.Sprintf("https://github.com/%s/%s", gitOwner, repo), "-b", fmt.Sprintf("v%s", r.Git.TagName))
	repo := cv.repo()
	cmd := exec.Command("/usr/bin/git", "clone", cv.repo())
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to clone: %v (%s)", err, out)
	}

	repoRoot := filepath.Join(dir, repo[strings.LastIndex(repo, "/"):])

	cmd = exec.Command("/usr/bin/git", "reset", "--hard", r.Git.CommitFingerprint)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to reset to commit: %v (%s)", err, out)
	}

	// Confirm that the git revision matches the manifest
	// This has been left in with the expectation that we check out by TagName above
	// at some point. If we don't do that then this is pretty redundant.
	cmd = exec.Command("/usr/bin/git", "rev-parse", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get HEAD revision: %v (%s)", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), r.Git.CommitFingerprint; got != want {
		return false, fmt.Errorf("expected revision %q but got %q for tag %q", want, got, r.Git.TagName)
	}

	// Make the elf file
	cmd = cv.makeCommand()
	cmd.Dir = repoRoot
	cmd.Env = append(cmd.Env, r.Build.Envs...)
	cmd.Env = append(cmd.Env, v.metadata.Envs...)
	cmd.Env = append(cmd.Env, v.tamago.Envs(r.Build.TamagoVersion)...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SEMVER_TAG=v%s", r.Git.TagName))
	klog.V(1).Infof("Running %q in %s", cmd.String(), repoRoot)
	if klog.V(2).Enabled() {
		for _, e := range cmd.Env {
			klog.V(2).Infof("  %s", e)
		}
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to make: %v (%s)", err, out)
	} else if klog.V(2).Enabled() {
		klog.V(2).Info(string(out))
	}

	// Hash the firmware artifact.
	data, err := os.ReadFile(filepath.Join(repoRoot, cv.binFile()))
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %v", cv.binFile(), err)
	}
	if got, want := sha256.Sum256(data), r.Output.FirmwareDigestSha256; !bytes.Equal(got[:], want) {
		// TODO: report this in a more visible way than an error in the log.
		klog.Errorf("Leaf index %d: ❌ failed to reproduce build %s@%s (%s) => (got %x, wanted %x)", i, r.Component, r.Git.TagName, r.Git.CommitFingerprint, got, want)
		return false, nil
	}

	klog.Infof("Leaf index %d: ✅ reproduced build %s@%s (%s) => %x", i, r.Component, r.Git.TagName, r.Git.CommitFingerprint, r.Output.FirmwareDigestSha256)
	cleanup = true
	return true, nil
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

type recoveryVerifier struct {
}

func (v recoveryVerifier) repo() string {
	// https://github.com/transparency-dev/armored-witness-boot/pull/89
	return "https://github.com/usbarmory/armory-ums"
}

func (v recoveryVerifier) makeCommand() *exec.Cmd {
	return exec.Command("/usr/bin/make", "imx")
}

func (v recoveryVerifier) binFile() string {
	return "armory-ums.imx"
}

func assertSigners(n *note.Note, names ...note.Verifier) error {
	needed := make(map[string]bool)
	for _, n := range names {
		needed[n.Name()] = true
	}
	for _, s := range n.Sigs {
		if !needed[s.Name] {
			return fmt.Errorf("unexpected sig for %s", s.Name)
		}
		delete(needed, s.Name)
	}
	if len(needed) > 0 {
		keys := make([]string, 0, len(needed))
		for k := range needed {
			keys = append(keys, k)
		}
		return fmt.Errorf("no sigs found for %v", keys)
	}
	return nil
}
