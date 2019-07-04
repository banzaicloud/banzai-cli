// Copyright © 2019 Banzai Cloud
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

package cluster

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type helmOptions struct {
	Context
}

func NewHelmCommand(banzaiCli cli.Cli) *cobra.Command {
	options := helmOptions{}

	cmd := &cobra.Command{
		Use:    "_helm",
		Hidden: true,
		Short:  "Wrapper to download and execute the Helm version matching the Tiller on the cluster of the current kubecontext, in a Helm home that is synchronized with Pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHelm(banzaiCli, options, args)
		},
	}
	options.Context = NewClusterContext(cmd, banzaiCli, "get")

	return cmd
}

func writeHelm(url, name string) error {

	tgz, err := http.Get(url)
	if err != nil {
		return emperror.Wrapf(err, "failed to download helm from %q", url)
	}
	defer tgz.Body.Close()

	zr, err := gzip.NewReader(tgz.Body)
	if err != nil {
		return emperror.Wrap(err, "failed to uncompress helm archive")
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			return emperror.Wrap(err, "failed to extract helm from archive")
		}
		if filepath.Base(hdr.Name) == "helm" {
			tempName := name + "~"
			f, err := os.OpenFile(tempName, (os.O_WRONLY | os.O_CREATE | os.O_EXCL), 0755)
			if err != nil {
				return emperror.Wrap(err, "failed to create temporary file for helm binary")
			}

			_, err = io.Copy(f, tr)
			f.Close()

			if err != nil {
				return emperror.Wrap(err, "failed to write helm binary")
			}

			return emperror.Wrap(os.Rename(tempName, name), "failed to move helm binary to its final place")
		}
	}
}

func runHelm(banzaiCli cli.Cli, options helmOptions, args []string) error {
	c := exec.Command("kubectl", "get", "deployment", "-n", "kube-system", "-o", "jsonpath={.items[0].spec.template.spec.containers[0].image}", "-l", "app=helm")
	out, err := c.Output()
	if err != nil {
		return emperror.Wrap(err, "failed to determine version of Tiller on the cluster")
	}
	parts := strings.Split(string(out), ":")
	if len(parts) != 2 {
		return errors.Errorf("failed to parse Tiller image name %q", out)
	}
	version := parts[1] // TODO check format

	// TODO use dir from config
	home, err := homedir.Dir()
	if err != nil {
		return emperror.Wrap(err, "can't compose the default config file path")
	}

	bindir := path.Join(home, ".banzai/bin")
	if err := os.MkdirAll(bindir, 0755); err != nil {
		return emperror.Wrapf(err, "failed to create %q directory", bindir)
	}

	url := fmt.Sprintf("https://get.helm.sh/helm-%s-%s-amd64.tar.gz", version, runtime.GOOS)
	name := path.Join(bindir, fmt.Sprintf("helm-%s", version))

	if _, err := os.Stat(name); err != nil {
		log.Infof("Downloading helm %s...", version)
		if runtime.GOARCH != "amd64" {
			return errors.Errorf("unsupported architecture: %v", runtime.GOARCH)
		}
		if err := writeHelm(url, name); err != nil {
			return emperror.Wrap(err, "failed to download helm client")
		}
		log.Infof("Helm %s downloaded successfully", version)
	}

	return nil
}
