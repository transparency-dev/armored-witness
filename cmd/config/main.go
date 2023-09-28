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
// The config tool builds serialised configs containing proof bundles.
// This is primarily useful for development work.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/transparency-dev/armored-witness-common/release/firmware"
	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"github.com/transparency-dev/armored-witness-common/release/firmware/update"
	"github.com/transparency-dev/merkle/rfc6962"
	"github.com/transparency-dev/serverless-log/client"
	"golang.org/x/mod/sumdb/note"
	"k8s.io/klog"
)

var (
	outputFile         = flag.String("output_file", "", "File to write the bundle to.")
	logBaseURL         = flag.String("log_url", "", "Base URL for the firmware transparency log to use.")
	logOrigin          = flag.String("log_origin", "", "FT log origin string")
	logPubKeyFile      = flag.String("log_pubkey_file", "", "File containing the FT log's public key in Note verifier format.")
	binaryBaseURL      = flag.String("bin_url", "", "Base URL for fetching firmware artefacts.")
	manifestFile       = flag.String("manifest_file", "", "Manifest to build a bundle for.")
	manifestPubKeyFile = flag.String("manifest_pubkey_file", "", "File containing a Note verifier string to verify manifest signatures.")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	mv := verifierOrDie(*manifestPubKeyFile)
	manifest, release := loadManifestOrDie(*manifestFile, mv)

	binFetcher := binFetcherOrDir(*binaryBaseURL)
	fwBin, err := binFetcher(ctx, release)
	if err != nil {
		klog.Exitf("Failed to fetch binary for manifest: %v", err)
	}

	logFetcher := newFetcherOrDie(*logBaseURL)
	logHasher := rfc6962.DefaultHasher
	lst, err := client.NewLogStateTracker(
		ctx,
		logFetcher,
		logHasher,
		nil,
		verifierOrDie(*logPubKeyFile),
		*logOrigin,
		client.UnilateralConsensus(logFetcher),
	)
	if _, _, _, err := lst.Update(ctx); err != nil {
		klog.Exitf("Update: %v", err)
	}

	idx, err := client.LookupIndex(ctx, logFetcher, logHasher.HashLeaf(manifest))
	if err != nil {
		klog.Exitf("LookupIndex: %v", err)
	}
	klog.Infof("Found manifest at index %d", idx)

	incP, err := lst.ProofBuilder.InclusionProof(ctx, idx)
	if err != nil {
		klog.Exitf("InclusionProof: %v", err)
	}

	bundle := firmware.Bundle{
		Checkpoint:     lst.LatestConsistentRaw,
		Index:          idx,
		InclusionProof: incP,
		Manifest:       manifest,
		Firmware:       fwBin,
	}
	// TODO: firmware.NewBundleVerifier()

	// We don't want the firmware in the encoded config, we only
	// needed it to verify the bundle above.
	bundle.Firmware = nil
	jsn, _ := json.MarshalIndent(&bundle, "", " ")
	klog.Infof("ProofBundle:\n%s", string(jsn))

	config := firmware.Config{
		Bundle: bundle,
	}
	configGob, err := config.Encode()
	if err != nil {
		klog.Exitf("config.Encode: %v", err)
	}

	if err := os.WriteFile(*outputFile, configGob, 0o644); err != nil {
		klog.Exitf("WriteFile: %v", err)
	}

	klog.Infof("Wrote %d bytes t chof config+bundle to %q", len(configGob), *outputFile)
}

// newFetcherOrDie creates a Fetcher for the log at the given root location.
func newFetcherOrDie(logURL string) client.Fetcher {
	root, err := url.Parse(logURL)
	if err != nil {
		klog.Exitf("Couldn't parse log_base_url: %v", err)
	}

	get := getByScheme[root.Scheme]
	if get == nil {
		klog.Exitf("Unsupported URL scheme %s", root.Scheme)
	}

	r := func(ctx context.Context, p string) ([]byte, error) {
		u, err := root.Parse(p)
		if err != nil {
			return nil, err
		}
		return get(ctx, u)
	}
	return r
}

var getByScheme = map[string]func(context.Context, *url.URL) ([]byte, error){
	"http":  readHTTP,
	"https": readHTTP,
	"file": func(_ context.Context, u *url.URL) ([]byte, error) {
		return os.ReadFile(u.Path)
	},
}

func readHTTP(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 404:
		klog.Infof("Not found: %q", u.String())
		return nil, os.ErrNotExist
	case 200:
		break
	default:
		return nil, fmt.Errorf("unexpected http status %q", resp.Status)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Errorf("resp.Body.Close(): %v", err)
		}
	}()
	return io.ReadAll(resp.Body)
}

func verifierOrDie(p string) note.Verifier {
	vs, err := os.ReadFile(p)
	if err != nil {
		klog.Exitf("Failed to read pub key file %q: %v", p, err)
	}
	v, err := note.NewVerifier(string(vs))
	if err != nil {
		klog.Exitf("Invalid note verifier string %q: %v", vs, err)
	}
	return v
}

func binFetcherOrDir(binURL string) func(context.Context, ftlog.FirmwareRelease) ([]byte, error) {
	f := newFetcherOrDie(binURL)

	return func(ctx context.Context, fr ftlog.FirmwareRelease) ([]byte, error) {
		p, err := update.BinaryPath(fr)
		if err != nil {
			return nil, err
		}

		return f(ctx, p)
	}
}

func loadManifestOrDie(p string, v note.Verifier) ([]byte, ftlog.FirmwareRelease) {
	b, err := os.ReadFile(p)
	if err != nil {
		klog.Exitf("Failed to read manifest %q: %v", p, err)
	}
	n, err := note.Open(b, note.VerifierList(v))
	if err != nil {
		klog.Exitf("Failed to verify manifest: %v", err)
	}
	var fr ftlog.FirmwareRelease
	if err := json.Unmarshal([]byte(n.Text), &fr); err != nil {
		klog.Exitf("Invalid manifest contents %q: %v", n.Text, err)
	}
	return b, fr
}
