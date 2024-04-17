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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/transparency-dev/armored-witness/internal/release"
	"golang.org/x/exp/maps"
	"k8s.io/klog/v2"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if tName, err := cmd.Flags().GetString("template"); err != nil {
			klog.Exitf("Failed to get `template` flag: %v", err)
		} else if tName != "" {
			applyFlagTemplate(cmd, tName)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("template", "prod", fmt.Sprintf("One of %v", maps.Keys(release.Templates)))
	rootCmd.PersistentFlags().String("log_url", "", "URL identifying the location of the log.")
	rootCmd.PersistentFlags().String("log_origin", "", "The expected first line of checkpoints issued by the log.")
	rootCmd.PersistentFlags().String("log_pubkey", "", "The log's public key.")
	rootCmd.PersistentFlags().String("os_release_pubkey1", "", "The first OS release signer's public key.")
	rootCmd.PersistentFlags().String("os_release_pubkey2", "", "The second OS release signer's public key.")
	rootCmd.PersistentFlags().String("applet_release_pubkey", "", "The applet release signer's public key.")
	rootCmd.PersistentFlags().String("boot_release_pubkey", "", "The boot release signer's public key.")
	rootCmd.PersistentFlags().String("recovery_release_pubkey", "", "The recovery release signer's public key.")

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

func applyFlagTemplate(cmd *cobra.Command, k string) {
	klog.Infof("Using template flags %q", k)
	tmpl, ok := release.Templates[k]
	if !ok {
		klog.Exitf("No such template %q", k)
	}
	// Define which flags we need, and map them to what they're called in the template.
	flagMap := map[string]string{
		"log_url":                 "firmware_log_url",
		"log_origin":              "firmware_log_origin",
		"log_pubkey":              "firmware_log_verifier",
		"os_release_pubkey1":      "os_verifier_1",
		"os_release_pubkey2":      "os_verifier_2",
		"applet_release_pubkey":   "applet_verifier",
		"boot_release_pubkey":     "boot_verifier",
		"recovery_release_pubkey": "recovery_verifier",
	}
	for f, m := range flagMap {
		v, ok := tmpl[m]
		if !ok {
			klog.Exitf("Unknown template flag %q", m)
		}
		if c, err := cmd.Flags().GetString(f); err != nil {
			klog.Exitf("Internal error applying template: %v", err)
		} else if c != "" {
			klog.Exitf("Cannot set --template and --%s", f)
		}
		klog.Infof("Using template flag setting --%v=%v", f, v)
		if err := cmd.Flags().Set(f, v); err != nil {
			klog.Exitf("Internal error setting templated flag %q: %v", f, err)

		}
	}
}
