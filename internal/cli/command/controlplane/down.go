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

package controlplane

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type destroyOptions struct {
	controlPlaneInstallerOptions
}

// NewDownCommand creates a new cobra.Command for `banzai controlplane down`.
func NewDownCommand(banzaiCli cli.Cli) *cobra.Command {
	options := destroyOptions{}

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy the controlplane",
		Long:  "Destroy a controlplane based on json stdin or interactive session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroy(options, banzaiCli)
		},
	}

	flags := cmd.Flags()

	bindInstallerFlags(flags, &options.controlPlaneInstallerOptions)

	return cmd
}

func runDestroy(options destroyOptions, banzaiCli cli.Cli) error {

	if isInteractive() {
		var destroy bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to DESTROY the controlplane now?",
				Default: true,
			},
			&destroy,
			nil,
		)

		if !destroy {
			return errors.New("controlplane destroy cancelled")
		}
	}

	// create temp dir for the files to attach
	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		return emperror.Wrapf(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(dir)

	kubeconfigName, err := filepath.Abs(filepath.Join(dir, "kubeconfig"))
	if err != nil {
		return emperror.Wrap(err, "failed to construct kubeconfig file name")
	}

	valuesName, err := filepath.Abs(valuesDefault)
	if err != nil {
		return emperror.Wrap(err, "failed to construct values file name")
	}

	var values map[string]interface{}
	rawValues, err := ioutil.ReadFile(valuesName)
	if err != nil {
		return emperror.Wrap(err, "failed to read control plane descriptor")
	}

	if err := utils.Unmarshal(rawValues, &values); err != nil {
		return emperror.Wrap(err, "failed to parse control plane descriptor")
	}

	kindCluster := isKINDClusterRequested(values)

	if err := copyKubeconfig(banzaiCli, kubeconfigName, kindCluster); err != nil {
		return emperror.Wrap(err, "failed to copy Kubeconfig")
	}

	tfdir, err := filepath.Abs("./.tfstate")
	if err != nil {
		return emperror.Wrap(err, "failed to construct tfstate directory path")
	}

	log.Info("controlplane is being destroyed")
	err = runInternal("destroy", valuesName, kubeconfigName, tfdir, options.controlPlaneInstallerOptions)
	if err != nil {
		return emperror.Wrap(err, "control plane destroy failed")
	}

	if kindCluster {
		err = deleteKINDCluster(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "KIND cluster destroy failed")
		}
	}

	return nil
}
