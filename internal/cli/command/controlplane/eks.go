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

const (
	eksModule = "module.eks"
)

func ensureEKSCluster(banzaiCli cli.Cli, options *cpContext, creds map[string]string) error {
	if options.kubeconfigExists() {
		return nil
	}

	log.Info("Creating Amazon EKS Kubernetes cluster...")
	if err := runTerraform("apply", options, banzaiCli, creds, eksModule, "local_file.eks_k8s_config", "local_file.eks_map_auth"); err != nil {
		return errors.WrapIf(err, "failed to create Amazon EKS Kubernetes cluster")
	}

	config, err := options.readEksK8sConfig()
	if err != nil {
		return errors.WrapIf(err, "failed to read Amazon EKS kubeconfig")
	}

	return options.writeKubeconfig(config)
}

func deleteEKSCluster(banzaiCli cli.Cli, options *cpContext, creds map[string]string) error {
	log.Info("Deleting Kubernetes cluster on Amazon EKS...")
	if err := runTerraform("destroy", options, banzaiCli, creds, eksModule); err != nil {
		return errors.WrapIf(err, "failed to delete Amazon EKS infrastructure")
	}

	return nil
}
