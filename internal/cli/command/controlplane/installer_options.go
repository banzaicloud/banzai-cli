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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	workspaceKey            = "installer.workspace"
	valuesFilename          = "values.yaml"
	kubeconfigFilename      = "kubeconfig"
	ec2HostFilename         = "ec2-host"
	sshkeyFilename          = "id_rsa"
	traefikAddressFilename  = "traefik-address"
	externalAddressFilename = "external-address"
	tfstateFilename         = "terraform.tfstate"
)

type cpContext struct {
	installerTag       string
	installerImageRepo string
	containerRuntime   string
	refreshState       bool
	pullInstaller      bool
	autoApprove        bool
	workspace          string
	banzaiCli          cli.Cli
}

func NewContext(cmd *cobra.Command, banzaiCli cli.Cli) *cpContext {
	ctx := cpContext{
		banzaiCli: banzaiCli,
	}

	flags := cmd.Flags()
	flags.StringVar(&ctx.installerTag, "image-tag", "latest", "Tag of installer Docker image to use")
	flags.StringVar(&ctx.installerImageRepo, "image", "docker.io/banzaicloud/cp-installer", "Name of Docker image repository to use")
	flags.BoolVar(&ctx.pullInstaller, "image-pull", true, "Pull installer image even if it's present locally")
	flags.BoolVar(&ctx.autoApprove, "auto-approve", true, "Automatically approve the changes to deploy")
	flags.StringVar(&ctx.workspace, "workspace", "", "Name of directory for storing the applied configuration and deployment status")
	flags.StringVar(&ctx.containerRuntime, "container-runtime", "docker", `Run the terraform command with "docker", "containerd" (crictl) or "exec" (execute locally)`)
	flags.BoolVar(&ctx.refreshState, "refresh-state", true, "Refresh terraform state for each run (turn off to save time during development)")
	flags.MarkHidden("refresh-state")
	return &ctx
}

func (c *cpContext) installerImage() string {
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
	return filepath.Join(c.workspace, kubeconfigFilename)
}

func (c *cpContext) kubeconfigExists() bool {
	_, err := os.Stat(c.kubeconfigPath())
	return err == nil
}

func (c *cpContext) tfstatePath() string {
	return filepath.Join(c.workspace, tfstateFilename)
}

func (c *cpContext) deleteTfstate() error {
	return os.Remove(c.tfstatePath())
}

func (c *cpContext) writeKubeconfig(outBytes []byte) error {
	path := c.kubeconfigPath()
	log.Debugf("writing kubeconfig file to %q", path)
	return errors.WrapIf(ioutil.WriteFile(path, outBytes, 0600), "failed to write kubeconfig file")
}

func (c *cpContext) deleteKubeconfig() error {
	return os.Remove(c.kubeconfigPath())
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

// Init completes the cp context from the options, env vars, and if possible from the user
func (c *cpContext) Init() error {
	if c.workspace == "" {
		c.workspace = viper.GetString(workspaceKey)
	}

	if c.workspace == "" {
		c.workspace = "default"
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

	err = os.MkdirAll(c.workspace, 0700)
	return errors.WrapIff(err, "failed to use %q as workspace path", c.workspace)
}
