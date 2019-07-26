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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
)

func ensureEC2Cluster(banzaiCli cli.Cli, options cpContext, values map[string]interface{}) error {
	if options.kubeconfigExists() {
		return nil
	}

	_, creds, err := input.GetAmazonCredentials()
	if err != nil {
		return emperror.Wrap(err, "failed to get AWS credentials")
	}

	log.Info("Creating Kubernetes cluster on AWS...")
	if err := runInternal("apply-infra", options, creds); err != nil {
		return emperror.Wrap(err, "failed to create AWS infrastructure")
	}

	hostBytes, err := ioutil.ReadFile(filepath.Join(options.workspace, "ec2-host"))
	if err != nil {
		return emperror.Wrap(err, "can't read host name of EC2 instance created")
	}
	host := strings.Trim(string(hostBytes), "\n")

	log.Info("retrieve kubernetes config from cluster")
	cmd := exec.Command("ssh", "-l", "centos", "-i", filepath.Join(options.workspace, ".ssh/id_rsa"), host, "sudo", "cat", "/etc/kubernetes/admin.conf")
	cmd.Stderr = os.Stderr
	config, err := cmd.Output()
	if err != nil {
		return emperror.Wrap(err, "failed to retrieve kubernetes config from cluster")
	}

	return options.writeKubeconfig(config)
}

func destroyEC2Cluster(banzaiCli cli.Cli, options cpContext) error {
	_, creds, err := input.GetAmazonCredentials()
	if err != nil {
		return emperror.Wrap(err, "failed to get AWS credentials")
	}

	log.Info("Destroying Kubernetes cluster on AWS...")
	err = runInternal("destroy-infra", options, creds)
	return emperror.Wrap(err, "failed to destroy AWS infrastructure")
}
