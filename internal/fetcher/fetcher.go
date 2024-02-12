// Copyright 2023 The Armored Witness authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/transparency-dev/armored-witness-common/release/firmware/ftlog"
	"github.com/transparency-dev/armored-witness-common/release/firmware/update"
	"github.com/transparency-dev/serverless-log/client"
	"k8s.io/klog/v2"
)

// New creates a Fetcher for the log at the given root location.
func New(root *url.URL) client.Fetcher {
	get := getByScheme[root.Scheme]
	if get == nil {
		panic(fmt.Errorf("unsupported URL scheme %s", root.Scheme))
	}

	return func(ctx context.Context, p string) ([]byte, error) {
		u, err := root.Parse(p)
		if err != nil {
			return nil, err
		}
		return get(ctx, u)
	}
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
	case http.StatusNotFound:
		klog.Infof("Not found: %q", u.String())
		return nil, os.ErrNotExist
	case http.StatusOK:
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

func BinaryFetcher(f client.Fetcher) func(context.Context, ftlog.FirmwareRelease) ([]byte, []byte, error) {
	return func(ctx context.Context, r ftlog.FirmwareRelease) ([]byte, []byte, error) {
		p, err := update.BinaryPath(r)
		if err != nil {
			return nil, nil, fmt.Errorf("BinaryPath: %v", err)
		}
		klog.Infof("Fetching %v bin from %q", r.Component, p)
		var b, s []byte
		if b, err = f(ctx, p); err != nil {
			return nil, nil, fmt.Errorf("failed to get %v binary from %q: %v", r.Component, p, err)
		}
		if len(r.HAB.SignatureDigestSha256) != 0 {
			if p, err = update.HABSignaturePath(r); err != nil {
				return nil, nil, fmt.Errorf("HABSignaturePath: %v", err)
			}
			b, err = f(ctx, p)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get %v HAB signature from %q: %v", r.Component, p, err)
			}
		}
		return b, s, nil
	}
}
