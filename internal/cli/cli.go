// Copyright © 2018 Banzai Cloud
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

package cli

import (
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"io"
	"os"
	"path"
	"sync"

	"github.com/goph/emperror"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const orgIdKey = "organization.id"

type Cli interface {
	Out() io.Writer
	Color() bool
	Client() *pipeline.APIClient
	Context() Context
}

type Context interface {
	OrganizationID() int32
	SetOrganizationID(id int32)
}

type banzaiCli struct {
	out        io.Writer
	ctx        Context
	client     *pipeline.APIClient
	clientOnce sync.Once
}

func NewCli(out io.Writer) Cli {
	return &banzaiCli{
		out: out,
		ctx: &banzaiContext{},
	}
}

func (c *banzaiCli) Out() io.Writer {
	return c.out
}

func (c *banzaiCli) Color() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return !viper.GetBool("formatting.no-color")
	}

	return viper.GetBool("formatting.force-color")
}

func (c *banzaiCli) Interactive() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
		return !viper.GetBool("formatting.no-interactive")
	}

	return viper.GetBool("formatting.force-interactive")
}

func (c *banzaiCli) Client() *pipeline.APIClient {
	c.clientOnce.Do(func() {
		config := pipeline.NewConfiguration()
		config.BasePath = viper.GetString("pipeline.basepath")
		config.UserAgent = "banzai-cli/1.0.0/go"
		config.HTTPClient = oauth2.NewClient(nil, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: viper.GetString("pipeline.token")},
		))

		c.client = pipeline.NewAPIClient(config)
	})

	return c.client
}

func (c *banzaiCli) Context() Context {
	return c.ctx
}

type banzaiContext struct{}

func (c *banzaiContext) OrganizationID() int32 {
	return viper.GetInt32(orgIdKey)
}

func (c *banzaiContext) SetOrganizationID(id int32) {
	viper.Set(orgIdKey, id)

	c.save()
}

func (c *banzaiContext) save() {
	log.Debug("writing config")

	if err := viper.WriteConfig(); err != nil {
		log.Infof("failed to write config: %v", err)
		home, _ := homedir.Dir()
		configPath := path.Join(home, ".banzai")
		os.MkdirAll(configPath, os.ModePerm)
		configPath = path.Join(configPath, "config.yaml")
		if err := viper.WriteConfigAs(configPath); err != nil {
			log.Fatal(emperror.Wrap(err, "failed to write config"))
		} else {
			log.Infof("config created at %v", configPath)
		}
	}
}
