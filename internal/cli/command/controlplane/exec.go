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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
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

	cmdEnv := map[string]string{}
	for k, v := range env {
		cmdEnv[k] = v
	}

	cmd := []string{"terraform"}
	for _, word := range strings.Split(command, " ") {
		cmd = append(cmd, word)
	}

	switch command {
	case "state list": // nop
	case "graph": // nop

	case "init":
		cmd = append(cmd, "-input=false", "-force-copy")

		if fileExists(filepath.Join(options.workspace, "state.tfvars")) {
			cmd = append(cmd, "-backend-config", "/workspace/state.tfvars")
		}

	case "apply":
		if options.AutoApprove() {
			cmd = append(cmd, "-auto-approve")
		}
		fallthrough

	default:
		cmd = append(cmd, []string{
			"-var", "workdir=/workspace",
			fmt.Sprintf("-refresh=%v", options.refreshState),
		}...)

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

func runContainer(command []string, options *cpContext, extraArgs []string, cmdOpt func(*exec.Cmd) error) error {
	args := []string{
		"run", "--rm", "--net-host",
		// fmt.Sprintf("--user=%d", os.Getuid()), // TODO
	}

	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
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

func runDocker(command []string, options *cpContext, extraArgs []string, cmdOpt func(*exec.Cmd) error) error {
	args := []string{
		"run", "--rm", "--net=host",
		fmt.Sprintf("--user=%d", os.Getuid()),
	}

	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
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

	img := options.installerImage()
	if !strings.Contains(img, "/") {
		log.Debugf("skip pulling local image %q", img)
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

	args = append(args, img)
	log.Info("Pulling Banzai Cloud Pipeline installer image...")

	log.Info(tool, " ", strings.Join(args, " "))

	cmd := exec.Command(tool, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func combineOutput(output string, err error) string {
	if err != nil {
		if output != "" {
			output += "\n\n"
		}
		output += err.Error()
	}
	return output
}

func runContainerCommand(options *cpContext, cmd []string, cmdEnv map[string]string) (string, error) {
	buffer := new(bytes.Buffer)
	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Stdout = buffer
		cmd.Stderr = buffer
		cmd.Env = os.Environ()
		for key, value := range cmdEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
		return nil
	}

	var err error
	switch options.containerRuntime {
	case runtimeExec:
		err = runLocally(cmd, cmdOpt)
	case runtimeDocker:
		args := []string{
			"-v", fmt.Sprintf("%s:/workspace", options.workspace),
		}
		for key := range cmdEnv {
			args = append(args, "-e", key)
		}
		err = runDocker(cmd, options, args, cmdOpt)
	case runtimeContainerd:
		args := []string{
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=/workspace,options=rbind:rw", options.workspace),
		}
		for key, value := range cmdEnv {
			args = append(args, "--env", fmt.Sprintf("%s=%s", key, value)) // env propagation does not work with ctr
		}
		err = runContainer(cmd, options, args, cmdOpt)
	default:
		err = errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}

	return buffer.String(), err
}

func readFilesFromContainerToMemory(options *cpContext, source string) (map[string][]byte, error) {
	// create gzipped archive (cz) and follow symlinks (h)
	cmd := []string{"sh", "-c", fmt.Sprintf("tar czh %s | base64", source)}

	var err error

	buffer := new(bytes.Buffer)
	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Stdout = buffer
		return nil
	}

	err = runContainerCommandGeneric(options, cmd, nil, cmdOpt)
	if err == nil {
		decoder := base64.NewDecoder(base64.StdEncoding, buffer)
		tarContents, err := untarInMemory(decoder)
		if err != nil {
			return nil, errors.WrapIf(err, "failed to untar source")
		}
		return tarContents, nil
	}
	return nil, errors.WrapIf(err, "failed to run container command")
}

func untarInMemory(reader io.Reader) (map[string][]byte, error) {
	zr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to uncompress helm archive")
	}
	defer zr.Close()

	contents := make(map[string][]byte)
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.WrapIf(err, "failed to extract next file from archive")
		}
		if hdr.FileInfo().Mode().IsRegular() {
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, tr)
			if err != nil {
				return nil, errors.WrapIff(err, "failed to read file contents from tar: %s", hdr.Name)
			}
			contents[hdr.Name] = buf.Bytes()
		}
	}
	return contents, nil
}

func runTerraformCommandGeneric(options *cpContext, cmd []string, cmdEnv map[string]string) error {
	logFile, err := options.createLog(cmd...)
	if err != nil {
		return errors.WrapIf(err, "failed to write output logs")
	}
	if logFile != nil {
		defer logFile.Close()
	}

	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Env = os.Environ()
		for key, value := range cmdEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Stdin = os.Stdin
		if options.containerRuntime == runtimeContainerd || logFile == nil {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = io.MultiWriter(logFile, os.Stdout)
			cmd.Stderr = io.MultiWriter(logFile, os.Stderr)
		}

		return nil
	}

	switch options.containerRuntime {
	case runtimeExec:
		return runLocally(cmd, cmdOpt)
	case runtimeDocker:
		args := []string{
			"-v", fmt.Sprintf("%s:/workspace", options.workspace),
			"-v", fmt.Sprintf("%s:/terraform/state.tf.json", options.workspace+"/state.tf.json"),
			"-v", fmt.Sprintf("%s:/terraform/.terraform/terraform.tfstate", options.workspace+"/.terraform/terraform.tfstate"),
		}
		if options.explicitState && options.tfstateExists() {
			args = append(args, "-v", fmt.Sprintf("%s:/terraform/terraform.tfstate", options.tfstatePath()))
		}
		if options.banzaiCli.Interactive() {
			args = append(args, "-ti")
		}
		for key := range cmdEnv {
			args = append(args, "-e", key)
		}
		return runDocker(cmd, options, args, cmdOpt)
	case runtimeContainerd:
		args := []string{
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=/workspace,options=rbind:rw", options.workspace),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=/terraform/state.tf.json,options=rbind:rw", options.workspace+"/state.tf.json"),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=/terraform/.terraform/terraform.tfstate,options=rbind:rw", options.workspace+"/.terraform/terraform.tfstate"),
		}
		if options.explicitState && options.tfstateExists() {
			args = append(args, "--mount", fmt.Sprintf("type=bind,src=%s,dst=/terraform/terraform.tfstate,options=rbind:rw", options.tfstatePath()))
		}
		if options.banzaiCli.Interactive() {
			args = append(args, "-t")
		}
		for key, value := range cmdEnv {
			args = append(args, "--env", fmt.Sprintf("%s=%s", key, value)) // env propagation does not work with ctr
		}

		var cmdLine string
		for _, word := range cmd {
			cmdLine += fmt.Sprintf("%q ", word)
		}
		cmdLine += "2>&1 | tee /workspace/.out"
		cmd = []string{"sh", "-o", "pipefail", "-c", cmdLine}

		containerErr := runContainer(cmd, options, args, cmdOpt)

		outpath := filepath.Join(options.workspace, ".out")
		if logFile != nil {
			f, err := os.Open(outpath)
			if err == nil {
				_, _ = io.Copy(logFile, f)
				_ = f.Close()
			}
		}
		_ = os.Remove(outpath)
		return containerErr
	default:
		return errors.Errorf("unknown container runtime: %q", options.containerRuntime)
	}
}

func runContainerCommandGeneric(options *cpContext, cmd []string, args []string, cmdOpt func(*exec.Cmd) error) error {
	switch options.containerRuntime {
	case runtimeExec:
		return runLocally(cmd, cmdOpt)
	case runtimeDocker:
		return runDocker(cmd, options, args, cmdOpt)
	case runtimeContainerd:
		return runContainer(cmd, options, args, cmdOpt)
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
