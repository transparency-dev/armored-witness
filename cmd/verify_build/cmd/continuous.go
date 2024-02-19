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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"github.com/transparency-dev/armored-witness/cmd/verify_build/cmd/internal/build"
	"github.com/transparency-dev/merkle/proof"
	"github.com/transparency-dev/merkle/rfc6962"
	"github.com/transparency-dev/serverless-log/client"
	"golang.org/x/mod/sumdb/note"
	"k8s.io/klog/v2"
)

// continuousCmd represents the continuous command
var (
	continuousCmd = &cobra.Command{
		Use:   "continuous",
		Short: "Continuously follow a log and verify all manifests that it commits to",
		Run:   continuous,
	}
)

func init() {
	rootCmd.AddCommand(continuousCmd)

	continuousCmd.Flags().String("log_url", "https://api.transparency.dev/armored-witness-firmware/ci/log/1/", "URL identifying the location of the log.")
	continuousCmd.Flags().String("log_origin", "transparency.dev/armored-witness/firmware_transparency/ci/1", "The expected first line of checkpoints issued by the log.")
	continuousCmd.Flags().String("log_pubkey", "transparency.dev-aw-ftlog-ci+f5479c1e+AR6gW0mycDtL17iM2uvQUThJsoiuSRirstEj9a5AdCCu", "The log's public key.")
	continuousCmd.Flags().String("os_release_pubkey1", "transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ", "The first OS release signer's public key.")
	continuousCmd.Flags().String("os_release_pubkey2", "transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh", "The second OS release signer's public key.")
	continuousCmd.Flags().String("applet_release_pubkey", "transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3", "The applet release signer's public key.")
	continuousCmd.Flags().String("boot_release_pubkey", "transparency.dev-aw-boot-ci+9f62b6ac+AbnipFmpRltfRiS9JCxLUcAZsbeH4noBOJXbVD3H5Eg4", "The boot release signer's public key.")
	continuousCmd.Flags().String("recovery_release_pubkey", "transparency.dev-aw-recovery-ci+cc699423+AarlJMSl0rbTMf31B5o9bqc6PHorwvF1GbwyJRXArbfg", "The recovery release signer's public key.")

	continuousCmd.Flags().Duration("poll_interval", 1*time.Minute, "The interval at which the log will be polled for new data.")
	continuousCmd.Flags().String("state_file", "", "File path for where checkpoints should be stored")
	continuousCmd.Flags().Bool("cleanup", true, "Set to false to keep git checkouts and make artifacts around after failed verification.")
	continuousCmd.Flags().String("tamago_dir", "/usr/local/tamago-go", "Directory in which versions of tamago should be installed to. User must have read/write permission to this directory.")
	continuousCmd.Flags().Uint64("start_index", 0, "Used for debugging to start verifying leaves from a given index. Only used if there is no prior checkpoint available.")
}

func continuous(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	st, isNew, err := stateTrackerFromFlags(ctx, cmd.Flags())
	if err != nil {
		klog.Exitf("Failed to create new LogStateTracker: %v", err)
	}

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

	monitor := Monitor{
		st:        st,
		stateFile: requireFlagString(cmd.Flags(), "state_file"),
		metadata:  metadata,
		handler:   rbv.VerifyManifest,
	}

	if isNew {
		startIndex := requireFlagUint64(cmd.Flags(), "start_index")
		klog.Infof("No previous checkpoint, starting at %d", startIndex)
		// This monitor has no memory of running before, so let's catch up with the log.
		if err := monitor.From(ctx, startIndex); err != nil {
			klog.Exitf("monitor.From(%d): %v", 0, err)
		}
	}

	klog.Infof("No known backlog, switching mode to poll log for new checkpoints. Current size: %d", st.LatestConsistent.Size)

	// We've processed all leaves committed to by the tracker's checkpoint, and now we enter polling mode.
	ticker := time.NewTicker(requireFlagDuration(cmd.Flags(), "poll_interval"))
	defer ticker.Stop()
	for {
		lastHead := st.LatestConsistent.Size
		if _, _, _, err := st.Update(ctx); err != nil {
			klog.Exitf("Failed to update checkpoint: %q", err)
		}
		if st.LatestConsistent.Size > lastHead {
			klog.V(1).Infof("Found new checkpoint for tree size %d, fetching new leaves", st.LatestConsistent.Size)
			if err := monitor.From(ctx, lastHead); err != nil {
				klog.Exitf("monitor.From(%d): %v", lastHead, err)
			}
		} else {
			klog.V(2).Infof("Polling: no new data found; tree size is still %d", st.LatestConsistent.Size)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Go around the loop again.
		}
	}
}

// Monitor verifiably checks inclusion of all leaves in a range, and then passes the
// parsed FirmwareRelease to a handler.
type Monitor struct {
	st        client.LogStateTracker
	stateFile string
	metadata  *build.ReleaseImplicitMetadata
	handler   func(context.Context, uint64, ftlog.FirmwareRelease) error
}

// From checks the leaves from `start` up to the checkpoint from the state tracker.
// Upon reaching the end of the leaves, the checkpoint is persisted in the state file.
func (m *Monitor) From(ctx context.Context, start uint64) error {
	fromCP := m.st.LatestConsistent
	pb, err := client.NewProofBuilder(ctx, fromCP, m.st.Hasher.HashChildren, m.st.Fetcher)
	if err != nil {
		return fmt.Errorf("failed to construct proof builder: %v", err)
	}
	klog.Infof("Running Monitor.From (%d, %d]", start, fromCP.Size)
	var resErr error
	for i := start; i < fromCP.Size; i++ {
		klog.V(1).Infof("Leaf index %d: fetching leaf", i)
		rawLeaf, err := client.GetLeaf(ctx, m.st.Fetcher, i)
		if err != nil {
			return fmt.Errorf("failed to get leaf at index %d: %v", i, err)
		}
		klog.V(1).Infof("Leaf index %d: fetching and verifying inclusion proof", i)
		hash := m.st.Hasher.HashLeaf(rawLeaf)
		ip, err := pb.InclusionProof(ctx, i)
		if err != nil {
			return fmt.Errorf("failed to get inclusion proof for index %d: %v", i, err)
		}

		if err := proof.VerifyInclusion(m.st.Hasher, i, fromCP.Size, hash, ip, fromCP.Hash); err != nil {
			return fmt.Errorf("VerifyInclusionProof() %d: %v", i, err)
		}

		klog.V(1).Infof("Leaf index %d: parsing", i)
		releaseNote, err := note.Open([]byte(rawLeaf), m.metadata.AllV)
		if err != nil {
			if e, ok := err.(*note.UnverifiedNoteError); ok && len(e.Note.UnverifiedSigs) > 0 {
				return fmt.Errorf("unknown signer %q for leaf at index %d: %v", e.Note.UnverifiedSigs[0].Name, i, err)
			}
			return fmt.Errorf("failed to open leaf note at index %d: %v", i, err)
		}

		var release ftlog.FirmwareRelease
		if err := json.Unmarshal([]byte(releaseNote.Text), &release); err != nil {
			return fmt.Errorf("failed to unmarshal release at index %d: %w", i, err)
		}

		switch release.Component {
		case ftlog.ComponentApplet:
			if err := assertSigners(releaseNote, m.metadata.AppV); err != nil {
				return fmt.Errorf("applet sig verification failed: %v", err)
			}
		case ftlog.ComponentOS:
			if err := assertSigners(releaseNote, m.metadata.OSV1, m.metadata.OSV2); err != nil {
				return fmt.Errorf("os sig verification failed: %v", err)
			}
		case ftlog.ComponentBoot:
			if err := assertSigners(releaseNote, m.metadata.BootV); err != nil {
				return fmt.Errorf("boot sig verification failed: %v", err)
			}
		case ftlog.ComponentRecovery:
			if err := assertSigners(releaseNote, m.metadata.RecoveryV); err != nil {
				return fmt.Errorf("recovery sig verification failed: %v", err)
			}
		default:
			// TODO(mhutchinson): support boot and recovery
			return fmt.Errorf("Unsupported component: %q", release.Component)
		}

		klog.V(1).Infof("Leaf index %d: verifying manifest: %s@%s (%s)", i, release.Component, release.GitTagName, release.GitCommitFingerprint)
		if err := m.handler(ctx, i, release); err != nil {
			resErr = err
			klog.Errorf("Error verifying index %d: %v", i, err)
		}
	}
	if resErr != nil {
		return resErr
	}
	return os.WriteFile(m.stateFile, m.st.LatestConsistentRaw, 0644)
}

// stateTrackerFromFlags constructs a state tracker based on the flags provided to the main invocation.
// The checkpoint returned will be the checkpoint representing this monitor's view of the log history.
// A boolean is returned that is true if the checkpoint was fetched from the log to initialize state.
func stateTrackerFromFlags(ctx context.Context, f *pflag.FlagSet) (client.LogStateTracker, bool, error) {
	stateFile := requireFlagString(f, "state_file")
	logURL := requireFlagString(f, "log_url")
	logOrigin := requireFlagString(f, "log_origin")
	logPubKey := requireFlagString(f, "log_pubkey")

	state, err := os.ReadFile(stateFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return client.LogStateTracker{}, false, fmt.Errorf("could not read state file %q: %w", stateFile, err)
		}
		klog.Infof("State file %q missing. Will trust first checkpoint received from log.", stateFile)
	}

	root, err := url.Parse(logURL)
	if err != nil {
		return client.LogStateTracker{}, false, fmt.Errorf("failed to parse log URL %q: %w", logURL, err)
	}
	fetcher, err := newFetcher(root)
	if err != nil {
		return client.LogStateTracker{}, false, fmt.Errorf("failed to create fetcher: %v", err)
	}

	lSigV, err := note.NewVerifier(logPubKey)
	if err != nil {
		return client.LogStateTracker{}, false, fmt.Errorf("unable to create new log signature verifier: %w", err)
	}

	lst, err := client.NewLogStateTracker(ctx, fetcher, rfc6962.DefaultHasher, state, lSigV, logOrigin, client.UnilateralConsensus(fetcher))
	return lst, state == nil, err
}

// newFetcher creates a Fetcher for the log at the given root location.
func newFetcher(root *url.URL) (client.Fetcher, error) {
	if s := root.Scheme; s != "http" && s != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %s", s)
	}

	return func(ctx context.Context, p string) ([]byte, error) {
		u, err := root.Parse(p)
		if err != nil {
			return nil, err
		}
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("got non-OK status code (%d) from %s", resp.StatusCode, u)
		}
		return body, nil
	}, nil
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

func requireFlagString(f *pflag.FlagSet, name string) string {
	v, err := f.GetString(name)
	if err != nil {
		log.Fatalf("Getting flag %v: %v", name, err)
	}
	if v == "" {
		log.Fatalf("Flag %v must be specified", name)
	}
	return v
}

func requireFlagBool(f *pflag.FlagSet, name string) bool {
	v, err := f.GetBool(name)
	if err != nil {
		log.Fatalf("Getting flag %v: %v", name, err)
	}
	return v
}

func requireFlagUint64(f *pflag.FlagSet, name string) uint64 {
	v, err := f.GetUint64(name)
	if err != nil {
		log.Fatalf("Getting flag %v: %v", name, err)
	}
	return v
}

func requireFlagDuration(f *pflag.FlagSet, name string) time.Duration {
	v, err := f.GetDuration(name)
	if err != nil {
		log.Fatalf("Getting flag %v: %v", name, err)
	}
	return v
}
