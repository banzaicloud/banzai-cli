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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	workspaceKey            = "installer.workspace"
	valuesFilename          = "values.yaml"
	ec2HostFilename         = "ec2-host"
	sshkeyFilename          = "id_rsa"
	traefikAddressFilename  = "traefik-address"
	externalAddressFilename = "external-address"
	tfstateFilename         = "terraform.tfstate"
	logsDir                 = "logs"
	defaultImage            = "docker.io/banzaicloud/pipeline-installer"
	latestTag               = "latest"
)

type cpContext struct {
	installerTag        string
	installerImageRepo  string
	containerRuntime    string
	refreshState        bool
	pullInstaller       bool
	autoApprove         bool
	explicitState       bool
	workspace           string
	banzaiCli           cli.Cli
	installerPulled     *sync.Once
	installerPullResult error
	flags               *pflag.FlagSet
	logOutput           bool
}

func NewContext(cmd *cobra.Command, banzaiCli cli.Cli) *cpContext {
	ctx := cpContext{
		banzaiCli:       banzaiCli,
		installerPulled: new(sync.Once),
	}

	ctx.flags = cmd.Flags()
	ctx.flags.StringVar(&ctx.installerTag, "image-tag", latestTag, "Tag of installer Docker image to use")
	ctx.flags.StringVar(&ctx.installerImageRepo, "image", defaultImage, "Name of Docker image repository to use")
	ctx.flags.BoolVar(&ctx.pullInstaller, "image-pull", true, "Pull installer image even if it's present locally")
	ctx.flags.BoolVar(&ctx.autoApprove, "auto-approve", true, "Automatically approve the changes to deploy")
	ctx.flags.BoolVar(&ctx.logOutput, "log-output", true, "Log output of terraform calls")
	ctx.flags.StringVar(&ctx.workspace, "workspace", "", "Name of directory for storing the applied configuration and deployment status")
	ctx.flags.StringVar(&ctx.containerRuntime, "container-runtime", "auto", `Run the terraform command with "docker", "containerd" (crictl) or "exec" (execute locally)`)
	ctx.flags.BoolVar(&ctx.refreshState, "refresh-state", true, "Refresh terraform state for each run (turn off to save time during development)")
	ctx.flags.MarkHidden("refresh-state")
	return &ctx
}

func (c *cpContext) ensureImagePulled() error {
	c.installerPulled.Do(func() { c.installerPullResult = pullImage(c, c.banzaiCli) })
	return c.installerPullResult
}

func (c *cpContext) AutoApprove() bool {
	if !c.flags.Changed("auto-approve") {
		out := make(map[interface{}]interface{})
		err := c.readValues(out)
		if err == nil {
			if installer, ok := out["installer"].(map[interface{}]interface{}); ok {
				if auto, ok := installer["autoApprove"].(bool); ok {
					return auto
				}
			}
		} else {
			log.Debugf("failed to read values file: %v", err)
		}
	}

	return c.autoApprove
}

func (c *cpContext) installerImage() string {
	if c.installerImageRepo == defaultImage && c.installerTag == latestTag {
		out := make(map[interface{}]interface{})
		err := c.readValues(out)
		if err == nil {
			if installer, ok := out["installer"].(map[interface{}]interface{}); ok {
				if image, ok := installer["image"].(string); ok && image != "" {
					return image
				}
			}
		} else {
			log.Debugf("failed to read values file: %v", err)
		}
	}

	return fmt.Sprintf("%s:%s", c.installerImageRepo, c.installerTag)
}

func (c *cpContext) valuesPath() string {
	return filepath.Join(c.workspace, valuesFilename)
}

func (c *cpContext) valuesExists() bool {
	_, err := os.Stat(c.valuesPath())
	return err == nil
}

func (c *cpContext) writeValues(out interface{}) error {
	outBytes, err := yaml.Marshal(out)
	if err != nil {
		return errors.WrapIf(err, "failed to marshal values file")
	}

	path := c.valuesPath()
	log.Debugf("writing values file to %q", path)
	return errors.WrapIf(ioutil.WriteFile(path, outBytes, 0600), "failed to write values file")
}

func (c *cpContext) readValues(out interface{}) error {
	path := c.valuesPath()
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.WrapIf(err, "failed to read values file")
	}

	return errors.WrapIf(yaml.Unmarshal(raw, out), "failed to parse values file")
}

func (c *cpContext) kubeconfigPath() string {
	return filepath.Join(c.workspace, ".kube", "config")
}

func (c *cpContext) legacyKubeconfigPath() string {
	return filepath.Join(c.workspace, "kubeconfig")
}

func (c *cpContext) kubeconfigExists() bool {
	return fileExists(c.kubeconfigPath()) || fileExists(c.legacyKubeconfigPath())
}

func (c *cpContext) tfstatePath() string {
	return filepath.Join(c.workspace, tfstateFilename)
}

func (c *cpContext) tfstateExists() bool {
	_, err := os.Stat(c.tfstatePath())
	return err == nil
}

func (c *cpContext) deleteTfstate() error {
	_, err := os.Stat(c.tfstatePath())
	if os.IsNotExist(err) {
		return nil
	}
	return os.Remove(c.tfstatePath())
}

func (c *cpContext) writeKubeconfig(outBytes []byte) error {
	path := c.kubeconfigPath()
	log.Debugf("writing kubeconfig file to %q", path)
	if ok, err := dirExists(filepath.Dir(path)); err != nil {
		return err
	} else if !ok {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return errors.WrapIf(err, "failed to create directories for kubeconfig")
		}
	}
	return errors.WrapIf(ioutil.WriteFile(path, outBytes, 0600), "failed to write kubeconfig file")
}

func (c *cpContext) deleteKubeconfig() error {
	if c.kubeconfigExists() {
		return os.Remove(c.kubeconfigPath())
	}
	return nil
}

func (c *cpContext) sshkeyPath() string {
	return filepath.Join(c.workspace, sshkeyFilename)
}

func (c *cpContext) traefikAddressPath() string {
	return filepath.Join(c.workspace, traefikAddressFilename)
}

func (c *cpContext) readTraefikAddress() (string, error) {
	path := c.traefikAddressPath()
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.WrapIf(err, "can't read endpoint URL")
	}
	return strings.Trim(string(bytes), "\n"), nil
}

func (c *cpContext) externalAddressPath() string {
	return filepath.Join(c.workspace, externalAddressFilename)
}

func (c *cpContext) readExternalAddress() (string, error) {
	path := c.externalAddressPath()
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.WrapIf(err, "can't read endpoint URL")
	}
	return strings.Trim(string(bytes), "\n"), nil
}

func (c *cpContext) ec2HostPath() string {
	return filepath.Join(c.workspace, ec2HostFilename)
}

func (c *cpContext) readEc2Host() (string, error) {
	path := c.ec2HostPath()
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.WrapIf(err, "can't read address of created EC2 instance")
	}
	return strings.Trim(string(bytes), "\n"), nil
}

func (c *cpContext) logDir() (string, error) {
	dir := filepath.Join(c.workspace, logsDir)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", errors.WrapIf(err, "failed to create log directory")
	}

	return dir, nil
}

func (c *cpContext) createLog(nameParts ...string) (io.WriteCloser, error) {
	if !c.logOutput {
		return nil, nil
	}

	logdir, err := c.logDir()
	if err != nil {
		return nil, err
	}

	ts, _ := time.Now().MarshalText()
	name := fmt.Sprintf("%s-%s.log", ts, strings.ReplaceAll(strings.Join(nameParts, "-"), "/", ""))
	return os.Create(filepath.Join(logdir, name))
}

func (c *cpContext) listLogs() (dir string, out []string, err error) {
	dir, err = c.logDir()
	if err != nil {
		return "", nil, err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to list logs")
	}

	for _, file := range files {
		out = append(out, file.Name())
	}
	return dir, out, nil
}

// Init completes the cp context from the options, env vars, and if possible from the user
func (c *cpContext) Init() error {
	if c.workspace == "" {
		c.workspace = viper.GetString(workspaceKey)
	}

	if c.workspace == "" {
		c.workspace = defaultWorkspace
	}

	if !strings.Contains(c.workspace, string(os.PathSeparator)) && c.workspace != "." && c.workspace != ".." {
		original := c.workspace
		c.workspace = filepath.Join(c.banzaiCli.Home(), "pipeline", c.workspace)
		if _, err := os.Stat(original); err == nil {
			original = "./" + original
			log.Warningf("Using workspace %q instead of %q. Pass --workspace=%q for relative path.", c.workspace, original, original)
		}
	}

	var err error
	c.workspace, err = filepath.Abs(c.workspace)
	if err != nil {
		return errors.WrapIff(err, "failed to calculate absolute path to %q", c.workspace)
	}

	log.Debugf("Using workspace %q.", c.workspace)

	switch c.containerRuntime {
	case "auto":
		if hasTool("docker") == nil {
			c.containerRuntime = runtimeDocker
		} else if hasTool("ctr") == nil || checkPKESupported() == nil {
			c.containerRuntime = runtimeContainerd
		} else {
			return errors.Errorf("neither docker, nor containerd is installed and working correctly on this machine")
		}
	case runtimeDocker:
		if err := hasTool("docker"); err != nil {
			return err
		}
	case runtimeContainerd:
		if err := hasTool("ctr"); err != nil {
			return err
		}
	case runtimeExec:
	default:
		return errors.Errorf("unknown container runtime: %q", c.containerRuntime)
	}

	err = os.MkdirAll(c.workspace, 0700)
	return errors.WrapIff(err, "failed to use %q as workspace path", c.workspace)
}
