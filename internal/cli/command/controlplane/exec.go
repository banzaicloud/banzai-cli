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

func runTerraform(command string, options *cpContext, banzaiCli cli.Cli, env map[string]string, targets ...string) error {
	cmdEnv := map[string]string{"KUBECONFIG": "/workspace/" + kubeconfigFilename}
	for k, v := range env {
		cmdEnv[k] = v
	}

	cmd := []string{"terraform",
		command,
		"-parallelism=1", // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271
		"-var", "workdir=/workspace",
		fmt.Sprintf("-refresh=%v", options.refreshState),
		"-state=/workspace/" + tfstateFilename,
	}

	if options.autoApprove {
		cmd = append(cmd, "-auto-approve")
	}

	for _, target := range targets {
		cmd = append(cmd, "-target", target)
	}

	switch options.containerRuntime {
	case "exec":
		return runLocally(cmd, cmdEnv)
	case "docker":
		return runDocker(cmd, options, banzaiCli, cmdEnv)
	case "containerd":
		return runContainer(cmd, options, banzaiCli, cmdEnv)
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}
}

// runLocally runs the given command locally (for development)
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

// runContainer runs the given installer command in the installer container with containerd (crictl)
func runContainer(command []string, options *cpContext, banzaiCli cli.Cli, env map[string]string) error {

	args := []string{
		"run", "--rm", "--net-host",
		// fmt.Sprintf("--user=%d", os.Getuid()), // TODO
		"--mount", fmt.Sprintf("type=bind,src=%s,dst=/workspace,options=rbind:rw", options.workspace),
	}

	if banzaiCli.Interactive() {
		args = append(args, "-t")
	}

	envs := os.Environ()
	for key, value := range env {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value)) // env propagation does not work with cri
	}

	args = append(append(args, options.installerImage(), "banzai-cp-installer"), command...)

	log.Info("ctr ", strings.Join(args, " "))

	cmd := exec.Command("ctr", args...)

	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}

// runDocker runs the given installer command in the installer docker container
func runDocker(command []string, options *cpContext, banzaiCli cli.Cli, env map[string]string) error {

	args := []string{
		"run", "--rm", "--net=host",
		fmt.Sprintf("--user=%d", os.Getuid()),
		"-v", fmt.Sprintf("%s:/workspace", options.workspace),
	}

	if banzaiCli.Interactive() {
		args = append(args, "-ti")
	}

	envs := os.Environ()
	for key, value := range env {
		args = append(args, "-e", key)
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	args = append(append(args, options.installerImage()), command...)

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}

func pullImage(options *cpContext, banzaiCli cli.Cli) error {
	if !options.pullInstaller {
		return nil
	}

	var args []string

	tool := options.containerRuntime
	switch options.containerRuntime {
	case "docker":
		args = []string{"pull"}
	case "containerd":
		tool = "ctr"
		args = []string{"image", "pull"}

	case "exec":
		return nil
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}

	args = append(args, options.installerImage())
	log.Info("Pulling Banzai Cloud Pipeline installer image...")

	log.Info(tool, " ", strings.Join(args, " "))

	cmd := exec.Command(tool, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
