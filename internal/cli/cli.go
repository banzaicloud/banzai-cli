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
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"

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
	Interactive() bool
	Client() *pipeline.APIClient
	HTTPTransport() *http.Transport
	Context() Context
	OutputFormat() string
	Home() string
}

type Context interface {
	OrganizationID() int32
	SetOrganizationID(id int32)
	SetToken(token string)
}

type banzaiCli struct {
	out        io.Writer
	client     *pipeline.APIClient
	clientOnce sync.Once
}

func NewCli(out io.Writer) Cli {
	return &banzaiCli{
		out: out,
	}
}

func (c *banzaiCli) Out() io.Writer {
	return c.out
}

func (c *banzaiCli) Home() string {
	// TODO use dir from config
	home, err := homedir.Dir()
	if err != nil {
		log.Errorf("failed to find home directory, falling back to /tmp: %v", err)
		home = "/tmp"
	}

	return filepath.Join(home, ".banzai")
}

func (c *banzaiCli) Color() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return !viper.GetBool("formatting.no-color")
	}

	return viper.GetBool("formatting.force-color")
}

func (c *banzaiCli) OutputFormat() string {
	return viper.GetString("output.format")
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

		config.HTTPClient.Transport.(*oauth2.Transport).Base = c.HTTPTransport()

		c.client = pipeline.NewAPIClient(config)
	})

	return c.client
}

func (c *banzaiCli) HTTPTransport() *http.Transport {
	skip := viper.GetBool("pipeline.tls-skip-verify")
	pemCerts := []byte(viper.GetString("pipeline.tls-ca-cert"))
	if caFile := viper.GetString("pipeline.tls-ca-file"); len(pemCerts) == 0 && caFile != "" {
		dat, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Errorf("failed to read CA certificate from %q: %v", caFile, err)
		} else {
			pemCerts = dat
		}
	}

	if skip || len(pemCerts) > 0 {
		tls := &tls.Config{
			InsecureSkipVerify: skip,
		}

		if len(pemCerts) > 0 {
			tls.RootCAs = x509.NewCertPool()
			ok := tls.RootCAs.AppendCertsFromPEM(pemCerts)
			if !ok {
				log.Error("failed to parse CA certificates")
			} else {
				log.Debugf("CA certs parsed (%d certs)", len(tls.RootCAs.Subjects()))
			}
		}

		return &http.Transport{
			TLSClientConfig: tls,
		}
	}

	return http.DefaultTransport.(*http.Transport)
}

func (c *banzaiCli) Context() Context {
	return c
}

func (c *banzaiCli) OrganizationID() int32 {
	return viper.GetInt32(orgIdKey)
}

func (c *banzaiCli) SetOrganizationID(id int32) {
	viper.Set(orgIdKey, id)

	c.save()
}

func (c *banzaiCli) SetToken(token string) {
	viper.Set("pipeline.token", token)

	c.save()
	c.clientOnce = sync.Once{}
}

func (c *banzaiCli) save() {
	log.Debug("writing config")

	if viper.ConfigFileUsed() == "" {
		log.Debug("no config file defined, falling back to default location $HOME/.banzai")

		home, _ := homedir.Dir()
		configPath := path.Join(home, ".banzai")
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			log.Fatal(emperror.Wrap(err, "failed to create config dir"))
		}

		configPath = filepath.Join(configPath, "config.yaml")
		err = viper.WriteConfigAs(configPath)
		if err != nil {
			log.Fatal(emperror.Wrap(err, "failed to write config"))
		}

		log.Infof("config created at %v", configPath)
		return
	}

	if _, err := os.Stat(filepath.Dir(viper.ConfigFileUsed())); os.IsNotExist(err) {
		log.Debug("creating config dir")

		configPath := filepath.Dir(viper.ConfigFileUsed())
		err := os.MkdirAll(configPath, 0700)
		if err != nil {
			log.Fatal(emperror.Wrap(err, "failed to create config dir"))
		}
	}

	err := viper.WriteConfig()
	if err != nil {
		log.Fatalf("failed to write config: %v", err)
	}
}
