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
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
)

const (
	defaultAwsRegion = "us-west-1"
	ec2Module        = "module.ec2"
)

func ensureEC2Cluster(banzaiCli cli.Cli, options cpContext, creds map[string]string, useGeneratedKey bool) error {
	if options.kubeconfigExists() {
		return nil
	}

	log.Info("Creating Kubernetes cluster on AWS...")
	if err := runTerraform("apply", options, banzaiCli, creds, ec2Module, "local_file.ec2_private_key_pem", "local_file.ec2_host"); err != nil {
		return errors.WrapIf(err, "failed to create AWS infrastructure")
	}

	host, err := options.readEc2Host()
	if err != nil {
		return err
	}

	log.Infof("retrieve kubernetes config from cluster %q", host)

	argv := []string{"-oStrictHostKeyChecking=no", "-l", "centos"}
	if useGeneratedKey {
		argv = append(argv, "-i", options.sshkeyPath(), "-F", "/dev/null")
	}

	argv = append(argv, host, "sudo", "cat", "/etc/kubernetes/admin.conf")

	log.Debugf("ssh %s", strings.Join(argv, " "))
	cmd := exec.Command("ssh", argv...)
	if !useGeneratedKey {
		cmd.Env = []string{"LC_ALL=C"} // prevent use of agent
	}

	cmd.Stderr = os.Stderr
	config, err := cmd.Output()
	if err != nil {
		return errors.WrapIf(err, "failed to retrieve kubernetes config from cluster")
	}

	return options.writeKubeconfig(config)
}

func deleteEC2Cluster(banzaiCli cli.Cli, options cpContext, creds map[string]string) error {
	log.Info("Deleting Kubernetes cluster on AWS...")
	if err := runTerraform("destroy", options, banzaiCli, creds, ec2Module); err != nil {
		return errors.WrapIf(err, "failed to delete AWS infrastructure")
	}

	return nil
}
