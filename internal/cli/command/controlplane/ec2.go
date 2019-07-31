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
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
)

func ensureEC2Cluster(_ cli.Cli, options cpContext, creds map[string]string) error {

	if options.kubeconfigExists() {
		return nil
	}

	log.Info("Creating Kubernetes cluster on AWS...")
	const ec2Host = "ec2-host"
	const idRsa = "id_rsa"
	argv := []string{"terraform", "apply",
		"-target", "module.aws_provider",
		"-var", "values_file=/root/" + valuesFilename,
		"-var", "instance_host_file=/root/" + ec2Host,
		"-var", "id_rsa_file=/root/" + idRsa}
	if err := runInstaller(argv, options, creds); err != nil {
		return emperror.Wrap(err, "failed to create AWS infrastructure")
	}

	hostBytes, err := ioutil.ReadFile(filepath.Join(options.workspace, ec2Host))
	if err != nil {
		return emperror.Wrap(err, "can't read host name of EC2 instance created")
	}
	host := strings.Trim(string(hostBytes), "\n")

	log.Infof("retrieve kubernetes config from cluster %q", host)
	cmd := exec.Command("ssh", "-oStrictHostKeyChecking=no", "-l", "centos", "-i", filepath.Join(options.workspace, idRsa), host, "sudo", "cat", "/etc/kubernetes/admin.conf")
	cmd.Stderr = os.Stderr
	config, err := cmd.Output()
	if err != nil {
		return emperror.Wrap(err, "failed to retrieve kubernetes config from cluster")
	}

	return options.writeKubeconfig(config)
}
