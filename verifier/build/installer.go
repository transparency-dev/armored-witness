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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coreos/go-semver/semver"
	"k8s.io/klog/v2"
)

func newTamago(dir string) (Tamago, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		klog.V(1).Infof("Creating new tamago install directory at %s", dir)
		if err := os.Mkdir(dir, os.ModeDir); err != nil {
			return Tamago{}, err
		}
	}
	return Tamago{
		dir: dir,
	}, nil
}

type Tamago struct {
	dir string
}

// Switch ensures that the named version of Tamago is installed, or fails with
// an error.
func (t Tamago) Switch(v semver.Version) error {
	vDir := filepath.Join(t.dir, v.String())
	if _, err := os.Stat(vDir); os.IsNotExist(err) {
		if err := t.install(v, vDir); err != nil {
			return fmt.Errorf("failed to install tamago %s: %v", v, err)
		}
	}
	goPath := filepath.Join(vDir, "bin", "go")
	if _, err := os.Stat(goPath); os.IsNotExist(err) {
		return fmt.Errorf("tamago version %q not available at %q: %v", v, goPath, err)
	}
	return nil
}

func (t Tamago) Envs(v semver.Version) []string {
	vDir := filepath.Join(t.dir, v.String())
	goPath := filepath.Join(vDir, "bin", "go")
	return []string{
		fmt.Sprintf("GOPATH=%s", vDir),
		fmt.Sprintf("GOCACHE=%s/go-cache", t.dir),
		fmt.Sprintf("TAMAGO=%s", goPath),
	}
}

func (t Tamago) install(v semver.Version, dir string) error {
	klog.Infof("Downloading and installing tamago %s", v)
	u := fmt.Sprintf("https://github.com/usbarmory/tamago-go/releases/download/tamago-go%s/tamago-go%s.linux-amd64.tar.gz", v, v)
	curl := exec.Command("curl", "-sfL", u)
	tar := exec.Command("tar", "-xzf", "-", "-C", dir, "--strip-components", "3")
	var err error
	tar.Stdin, err = curl.StdoutPipe()
	tar.Stdout = os.Stdout
	if err != nil {
		return err
	}
	if err := tar.Start(); err != nil {
		return err
	}
	// Create the directory and then extract into it
	if err := os.Mkdir(dir, os.ModeDir); err != nil {
		return err
	}
	if err := curl.Run(); err != nil {
		return err
	}
	if err := tar.Wait(); err != nil {
		return err
	}
	klog.Infof("Installed tamago %s at %s", v, dir)
	return nil
}
