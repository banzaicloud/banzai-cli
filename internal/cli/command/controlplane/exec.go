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
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils/untar"
	log "github.com/sirupsen/logrus"
)

const (
	runtimeDocker     = "docker"
	runtimeContainerd = "containerd"
	runtimeExec       = "exec"
)

func runTerraform(command string, options *cpContext, env map[string]string, targets ...string) error {
	err := options.ensureImagePulled()
	if err != nil {
		return errors.WrapIf(err, "failed to pull cp-installer")
	}

	cmdEnv := map[string]string{"KUBECONFIG": "/workspace/" + kubeconfigFilename}
	for k, v := range env {
		cmdEnv[k] = v
	}

	cmd := []string{"terraform", command}

	if command != "init" {
		cmd = append(cmd, []string{
			"-var", "workdir=/workspace",
			fmt.Sprintf("-refresh=%v", options.refreshState),
		}...)

		if options.AutoApprove() {
			cmd = append(cmd, "-auto-approve")
		}

		for _, target := range targets {
			cmd = append(cmd, "-target", target)
		}
	}

	return runTerraformCommandGeneric(options, cmd, cmdEnv)
}

// runLocally runs the given command locally (for development)
func runLocally(command []string, cmdOpt func(*exec.Cmd) error) error {
	log.Info(strings.Join(command, " "))

	cmd := exec.Command(command[0], command[1:]...)
	if cmdOpt != nil {
		if err := cmdOpt(cmd); err != nil {
			return errors.WrapIf(err, "failed to add optional command options")
		}
	}

	return errors.WithStack(cmd.Run())
}

func runContainer(command []string, options *cpContext, argOpt func(*[]string) error, cmdOpt func(*exec.Cmd) error) error {
	args := []string{
		"run", "--rm", "--net-host",
		// fmt.Sprintf("--user=%d", os.Getuid()), // TODO
	}

	if argOpt != nil {
		if err := argOpt(&args); err != nil {
			return errors.WrapIf(err, "failed to add optional arguments")
		}
	}

	args = append(append(args, options.installerImage(), "banzai-cp-installer"), command...)

	ctrCmd, err := lookupTool("ctr")
	if err != nil {
		return err
	}

	log.Info("ctr ", strings.Join(args, " "))

	cmd := exec.Command(ctrCmd, args...)

	if cmdOpt != nil {
		if err := cmdOpt(cmd); err != nil {
			return errors.WrapIf(err, "failed to add optional command options")
		}
	}

	return errors.WithStack(cmd.Run())
}

func runDocker(command []string, options *cpContext, argOpt func(*[]string) error, cmdOpt func(*exec.Cmd) error) error {
	args := []string{
		"run", "--rm", "--net=host",
		fmt.Sprintf("--user=%d", os.Getuid()),
	}

	if argOpt != nil {
		if err := argOpt(&args); err != nil {
			return errors.WrapIf(err, "failed to add optional arguments")
		}
	}

	args = append(append(args, options.installerImage()), command...)

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)
	if cmdOpt != nil {
		if err := cmdOpt(cmd); err != nil {
			return errors.WrapIf(err, "failed to add optional command options")
		}
	}

	return errors.WithStack(cmd.Run())
}

func pullImage(options *cpContext, _ cli.Cli) error {
	if !options.pullInstaller {
		return nil
	}

	var args []string

	tool := options.containerRuntime
	switch options.containerRuntime {
	case runtimeDocker:
		args = []string{"pull"}
	case runtimeContainerd:
		tool = "ctr"
		args = []string{"image", "pull"}

	case runtimeExec:
		return nil
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}

	tool, err := lookupTool(tool)
	if err != nil {
		return err
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

func exportFilesFromContainer(options *cpContext, source string, destination string) error {
	// create gzipped archive (cz) and follow symlinks (h)
	cmd := []string{"sh", "-c", fmt.Sprintf("tar czh %s | base64", source)}

	var err error

	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	buffer := new(bytes.Buffer)
	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Stdout = buffer
		return nil
	}

	err = runContainerCommandGeneric(options, cmd, nil, cmdOpt)
	if err == nil {
		decoder := base64.NewDecoder(base64.StdEncoding, buffer)
		if err := untar.Untar(decoder, destination); err != nil {
			return errors.Wrapf(err, "failed to untar container source %s to %s", source, destination)
		}
	}
	return errors.WrapIf(err, "failed to run container command")
}

func runTerraformCommandGeneric(options *cpContext, cmd []string, cmdEnv map[string]string) error {
	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Env = os.Environ()
		for key, value := range cmdEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return nil
	}

	switch options.containerRuntime {
	case runtimeExec:
		return runLocally(cmd, cmdOpt)
	case runtimeDocker:
		argOpt := func(args *[]string) error {
			*args = append(*args,
			"-v", fmt.Sprintf("%s:/workspace", options.workspace),
				"-v", fmt.Sprintf("%s:/terraform/state.tf", options.workspace+"/state.tf"),
				"-v", fmt.Sprintf("%s:/terraform/.terraform/terraform.tfstate", options.workspace+"/.terraform/terraform.tfstate"))
			if options.banzaiCli.Interactive() {
				*args = append(*args, "-ti")
			}
			for key, _ := range cmdEnv {
				*args = append(*args, "-e", key)
			}
			return nil
		}
		return runDocker(cmd, options, argOpt, cmdOpt)
	case runtimeContainerd:
		argOpt := func(args *[]string) error {
			*args = append(*args,
				"--mount", fmt.Sprintf("type=bind,src=%s,dst=/workspace,options=rbind:rw", options.workspace),
				"--mount", fmt.Sprintf("type=bind,src=%s,dst=/terraform/state.tf,options=rbind:rw", options.workspace+"/state.tf"),
				"--mount", fmt.Sprintf("type=bind,src=%s,dst=/terraform/.terraform/terraform.tfstate,options=rbind:rw", options.workspace+"/.terraform/terraform.tfstate"))
			if options.banzaiCli.Interactive() {
				*args = append(*args, "-t")
			}
			for key, value := range cmdEnv {
				*args = append(*args, "--env", fmt.Sprintf("%s=%s", key, value)) // env propagation does not work with ctr
			}
			return nil
		}
		return runContainer(cmd, options, argOpt, cmdOpt)
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}
}

func runContainerCommandGeneric(options *cpContext, cmd []string, argOpt func(*[]string) error, cmdOpt func(*exec.Cmd) error) error {
	switch options.containerRuntime {
	case runtimeExec:
		return runLocally(cmd, cmdOpt)
	case runtimeDocker:
		return runDocker(cmd, options, argOpt, cmdOpt)
	case runtimeContainerd:
		return runContainer(cmd, options, argOpt, cmdOpt)
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}
}

func lookupTool(tool string) (string, error) {
	cmd, err := exec.LookPath(tool)

	if err != nil {
		cmd, err := exec.LookPath(filepath.Join("/usr/local/bin", tool))
		if err == nil {
			return cmd, nil
		}
	}

	return cmd, errors.Wrapf(err, "can't find %s in PATH", tool)
}

func hasTool(tool string) error {
	tool, err := lookupTool(tool)
	if err != nil {
		return err
	}

	cmd := exec.Command(tool, "version")
	err = cmd.Run()

	// grab the last non-empty error line
	if err, ok := err.(*exec.ExitError); ok {
		var out string
		for _, line := range strings.Split(string(err.Stderr), "\n") {
			if line != "" {
				out = line
			}
		}
		if out != "" {
			return errors.Errorf("`%s version` failed: %s", tool, out)
		}
	}

	return errors.Wrapf(err, "%s check failed", tool)
}
