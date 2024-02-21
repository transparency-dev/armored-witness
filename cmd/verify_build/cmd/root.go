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

// Package cmd contains commands for the verify_build tool.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "verify_build",
	Short: "A tool for verifying reproducible builds from manifest files",
	Long: `verify_build is a tool for verifying reproducible builds from manifest files.

Manifest files contain important information about firmware releases, and
are intended to be stored in firmware transparency logs. This tool implements
a verifier that checks that the binary committed to by a manifest file can be
reproduced by building it again.`,
}

func init() {
	rootCmd.PersistentFlags().String("log_url", "https://api.transparency.dev/armored-witness-firmware/ci/log/1/", "URL identifying the location of the log.")
	rootCmd.PersistentFlags().String("log_origin", "transparency.dev/armored-witness/firmware_transparency/ci/1", "The expected first line of checkpoints issued by the log.")
	rootCmd.PersistentFlags().String("log_pubkey", "transparency.dev-aw-ftlog-ci+f5479c1e+AR6gW0mycDtL17iM2uvQUThJsoiuSRirstEj9a5AdCCu", "The log's public key.")
	rootCmd.PersistentFlags().String("os_release_pubkey1", "transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ", "The first OS release signer's public key.")
	rootCmd.PersistentFlags().String("os_release_pubkey2", "transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh", "The second OS release signer's public key.")
	rootCmd.PersistentFlags().String("applet_release_pubkey", "transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3", "The applet release signer's public key.")
	rootCmd.PersistentFlags().String("boot_release_pubkey", "transparency.dev-aw-boot-ci+9f62b6ac+AbnipFmpRltfRiS9JCxLUcAZsbeH4noBOJXbVD3H5Eg4", "The boot release signer's public key.")
	rootCmd.PersistentFlags().String("recovery_release_pubkey", "transparency.dev-aw-recovery-ci+cc699423+AarlJMSl0rbTMf31B5o9bqc6PHorwvF1GbwyJRXArbfg", "The recovery release signer's public key.")

	rootCmd.PersistentFlags().Bool("cleanup", true, "Set to false to keep git checkouts and make artifacts around after failed verification.")
	rootCmd.PersistentFlags().String("tamago_dir", "/usr/local/tamago-go", "Directory in which versions of tamago should be installed to. User must have read/write permission to this directory.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
