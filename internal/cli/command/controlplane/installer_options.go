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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	workspaceKey       = "installer.workspace"
	valuesFilename     = "values.yaml"
	kubeconfigFilename = "Kubeconfig"
)

type cpContext struct {
	installerTag  string
	pullInstaller bool
	workspace     string
	banzaiCli     cli.Cli
}

func (c *cpContext) pullDockerImage() error {

	args := []string{
		"pull",
		fmt.Sprintf("banzaicloud/cp-installer:%s", c.installerTag),
	}

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func NewContext(cmd *cobra.Command, banzaiCli cli.Cli) *cpContext {
	ctx := cpContext{
		banzaiCli: banzaiCli,
	}

	flags := cmd.Flags()
	flags.StringVar(&ctx.installerTag, "image-tag", "latest", "Tag of banzaicloud/cp-installer Docker image to use")
	flags.BoolVar(&ctx.pullInstaller, "image-pull", true, "Pull cp-installer image even if it's present locally")
	flags.StringVar(&ctx.workspace, "workspace", "", "Name of directory for storing the applied configuration and deployment status")
	return &ctx
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
		return emperror.Wrap(err, "failed to marshal values file")
	}

	path := c.valuesPath()
	log.Debugf("writing values file to %q", path)
	return emperror.Wrap(ioutil.WriteFile(path, outBytes, 0600), "failed to write values file")
}

func (c *cpContext) readValues(out interface{}) error {
	path := c.valuesPath()
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return emperror.Wrap(err, "failed to read values file")
	}

	return emperror.Wrap(yaml.Unmarshal(raw, out), "failed to parse values file")
}

func (c *cpContext) kubeconfigPath() string {
	return filepath.Join(c.workspace, kubeconfigFilename)
}

func (c *cpContext) kubeconfigExists() bool {
	_, err := os.Stat(c.kubeconfigPath())
	return err == nil
}

func (c *cpContext) writeKubeconfig(outBytes []byte) error {
	path := c.kubeconfigPath()
	log.Debugf("writing kubeconfig file to %q", path)
	return emperror.Wrap(ioutil.WriteFile(path, outBytes, 0600), "failed to write kubeconfig file")
}

// Init completes the cp context from the options, env vars, and if possible from the user
func (c *cpContext) Init() error {
	if c.workspace == "" {
		c.workspace = viper.GetString(workspaceKey)
	}

	if c.workspace == "" {
		c.workspace = "default"
	}

	if !strings.Contains(c.workspace, string(os.PathSeparator)) {
		c.workspace = filepath.Join(c.banzaiCli.Home(), "pipeline", c.workspace)
	}

	var err error
	c.workspace, err = filepath.Abs(c.workspace)
	if err != nil {
		return emperror.Wrapf(err, "failed to calculate absolute path to %q", c.workspace)
	}

	err = os.MkdirAll(c.workspace, 0700)
	return emperror.Wrapf(err, "failed to use %q as workspace path", c.workspace)
}
