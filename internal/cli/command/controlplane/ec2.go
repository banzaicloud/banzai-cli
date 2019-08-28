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
	"os"
	"os/exec"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
)

const defaultAwsRegion = "us-west-1"

func ensureEC2Cluster(_ cli.Cli, options cpContext, creds map[string]string) error {
	if options.kubeconfigExists() {
		return nil
	}

	log.Info("Creating Kubernetes cluster on AWS...")
	if err := runInternal("apply", options, creds, "module.ec2"); err != nil {
		return errors.WrapIf(err, "failed to create AWS infrastructure")
	}

	host, err := options.readEc2Host()
	if err != nil {
		return err
	}

	log.Infof("retrieve kubernetes config from cluster %q", host)
	cmd := exec.Command("ssh", "-oStrictHostKeyChecking=no", "-l", "centos", "-i", options.sshkeyPath(), host, "sudo", "cat", "/etc/kubernetes/admin.conf")
	cmd.Stderr = os.Stderr
	config, err := cmd.Output()
	if err != nil {
		return errors.WrapIf(err, "failed to retrieve kubernetes config from cluster")
	}

	return options.writeKubeconfig(config)
}

func deleteEC2Cluster(_ cli.Cli, options cpContext, creds map[string]string) error {
	log.Info("Deleting Kubernetes cluster on AWS...")
	argv := []string{"terraform", "destroy",
		"-target", "module.aws_provider",
	}

	if err := runInstaller(argv, options, creds); err != nil {
		return errors.WrapIf(err, "failed to delete AWS infrastructure")
	}

	return nil
}
