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
	"fmt"
	"os"
	"os/exec"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
)

func runTerraform(command string, options cpContext, banzaiCli cli.Cli, env map[string]string, targets ...string) error {
	cmdEnv := map[string]string{"KUBECONFIG": "/root/" + kubeconfigFilename}
	for k, v := range env {
		cmdEnv[k] = v
	}

	cmd := []string{"terraform",
		command,
		"-parallelism=1", // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271
		"-var", "workdir=/root",
		"-state=/root/" + tfstateFilename,
	}

	if options.autoApprove {
		cmd = append(cmd, "-auto-approve")
	}

	for _, target := range targets {
		cmd = append(cmd, "-target", target)
	}

	if options.runLocally {
		return runLocally(cmd, cmdEnv)
	}

	return runInstaller(cmd, options, banzaiCli, cmdEnv)
}

// runInstaller runs the given command locally (for development)
func runLocally(command []string, env map[string]string) error {
	log.Info(strings.Join(command, " "))

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return errors.WithStack(cmd.Run())
}

// runInstaller runs the given installer command in the installer docker container
func runInstaller(command []string, options cpContext, banzaiCli cli.Cli, env map[string]string) error {

	args := []string{
		"run", "--rm", "--net=host",
		"-v", fmt.Sprintf("%s:/root", options.workspace),
	}

	if banzaiCli.Interactive() {
		args = append(args, "-ti")
	}

	envs := os.Environ()
	for key, value := range env {
		args = append(args, "-e", key)
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	args = append(append(args,
		fmt.Sprintf("banzaicloud/cp-installer:%s", options.installerTag)),
		command...)

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}
