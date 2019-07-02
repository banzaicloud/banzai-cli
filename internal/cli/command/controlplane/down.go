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

	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type destroyOptions struct {
	controlPlaneInstallerOptions
}

// NewDownCommand creates a new cobra.Command for `banzai clontrolplane down`.
func NewDownCommand() *cobra.Command {
	options := destroyOptions{}

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy the controlplane",
		Long:  "Destroy a controlplane based on json stdin or interactive session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroy(options)
		},
	}

	flags := cmd.Flags()

	bindInstallerFlags(flags, &options.controlPlaneInstallerOptions)

	return cmd
}

func runDestroy(options destroyOptions) error {

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

	if err := copyKubeconfig(kubeconfigName); err != nil {
		return emperror.Wrap(err, "failed to copy Kubeconfig")
	}

	tfdir, err := filepath.Abs("./.tfstate")
	if err != nil {
		return emperror.Wrap(err, "failed to construct tfstate directory path")
	}

	valuesName, err := filepath.Abs(valuesDefault)
	if err != nil {
		return emperror.Wrap(err, "failed to construct values file name")
	}

	log.Info("controlplane is being destroyed")
	return emperror.Wrap(runInternal("destroy", valuesName, kubeconfigName, tfdir, options.controlPlaneInstallerOptions), "controlplane destroy failed")
}
