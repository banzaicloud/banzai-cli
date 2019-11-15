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
	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
)

func ensureCustomCluster(banzaiCli cli.Cli, options *cpContext, creds map[string]string) error {
	if options.kubeconfigExists() {
		return nil
	}

	log.Info("Creating custom infrastructure...")
	targets := []string{"module.custom", "local_file.k8s_config"}
	if err := runTerraform("apply", options, banzaiCli, creds, targets...); err != nil {
		return errors.WrapIf(err, "failed to create custom infrastructure")
	}

	return nil
}

func deleteCustomCluster(banzaiCli cli.Cli, options *cpContext, creds map[string]string) error {
	log.Info("Destroying custom infrastructure...")
	if err := runTerraform("destroy", options, banzaiCli, creds, "module.custom"); err != nil {
		return errors.WrapIf(err, "failed to destroy custom infrastructure")
	}

	return nil
}
