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

package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"
	"github.com/transparency-dev/armored-witness/cmd/verify_build/cmd/internal/build"
	"k8s.io/klog/v2"
)

// singleCmd represents the single-manifest command
var (
	singleCmd = &cobra.Command{
		Use:   "single",
		Short: "Verify a single manifest entry to check whether it can be reproducibly built",
		Long:  "The manifest to verify should be provided on stdin.",
		Run:   single,
	}
)

func init() {
	rootCmd.AddCommand(singleCmd)
}

func single(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	tamago, err := build.NewTamago(requireFlagString(cmd.Flags(), "tamago_dir"))
	if err != nil {
		klog.Exitf("Failed to init tamago: %v", err)
	}
	metadata, err := build.NewReleaseImplicitMetadata(
		requireFlagString(cmd.Flags(), "log_pubkey"),
		requireFlagString(cmd.Flags(), "os_release_pubkey1"),
		requireFlagString(cmd.Flags(), "os_release_pubkey2"),
		requireFlagString(cmd.Flags(), "applet_release_pubkey"),
		requireFlagString(cmd.Flags(), "boot_release_pubkey"),
		requireFlagString(cmd.Flags(), "recovery_release_pubkey"))
	if err != nil {
		klog.Exitf("Failed to initialize metadata: %v", err)
	}
	defer metadata.Cleanup()

	rbv, err := build.NewReproducibleBuildVerifier(requireFlagBool(cmd.Flags(), "cleanup"), tamago, metadata)
	if err != nil {
		klog.Exitf("Failed to create reproducible build verifier: %v", err)
	}

	nbs, err := io.ReadAll(cmd.Root().InOrStdin())
	if err != nil {
		klog.Exitf("Failed to read manifest from stdin: %v", err)
	}
	klog.V(1).Infof("Read %d bytes:\n%s", len(nbs), nbs)
	if err := rbv.Verify(ctx, 0, nbs); err != nil {
		klog.Exitf("Failed to verify manifest: %v", err)
	}
	klog.Info("Success!")
}
