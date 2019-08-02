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
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/banzaicloud/banzai-cli/.gen/cloudinfo"
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
	CloudinfoClient() *cloudinfo.APIClient
	Context() Context
	OutputFormat() string
	Home() string // Home is the path to the .banzai directory of the user
}

type Context interface {
	OrganizationID() int32
	SetOrganizationID(id int32)
	SetToken(token string)
	SetFingerprint(fingerprint string)
}

type banzaiCli struct {
	out                 io.Writer
	client              *pipeline.APIClient
	clientOnce          sync.Once
	cloudinfoClient     *cloudinfo.APIClient
	cloudinfoClientOnce sync.Once
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
	fingerprint := viper.GetString("pipeline.tls-fingerprint")
	fingerprintBytes, err := hex.DecodeString(fingerprint)
	if err != nil {
		log.Error(emperror.Wrapf(err, "invalid tls-fingerprint configuration %q", fingerprint))
		skip = false
	}

	pemCerts := []byte(viper.GetString("pipeline.tls-ca-cert"))
	if caFile := viper.GetString("pipeline.tls-ca-file"); len(pemCerts) == 0 && caFile != "" {
		dat, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Errorf("failed to read CA certificate from %q: %v", caFile, err)
		} else {
			pemCerts = dat
		}
	}

	/* #nosec G402 */
	if skip || len(pemCerts) > 0 || fingerprint != "" {
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

		if len(fingerprintBytes) != 0 {
			tls.VerifyPeerCertificate = makeFingerprintVerifier(fingerprintBytes)
		}

		return &http.Transport{
			TLSClientConfig: tls,
		}
	}

	return http.DefaultTransport.(*http.Transport)
}

func makeFingerprintVerifier(expected []byte) func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
		actual, err := getServerCertFingerprint(certificates)
		if err != nil {
			return err
		}
		if bytes.Compare(actual, expected) != 0 {
			return errors.Errorf("server certificate fingerprint %x does not match pinned value %x", actual, expected)
		}

		return nil
	}
}

// x509Error extracts the certificate validation errors from an url.Error or returns nil
func x509Error(err error) error {
	if err == nil {
		return nil
	}
	if err, ok := err.(*url.Error); ok {
		switch err.Err.(type) {
		case x509.UnknownAuthorityError, x509.CertificateInvalidError, x509.HostnameError:
			return err.Err
		}
	}
	return nil
}

// CheckPipelineEndpoint checks if the endpoint is a valid Pipeline endpoint.
// It returns the server cert's sha256 fingerprint if the endpoint is valid.
// If the endpoint is valid, but the TLS validatation failed, it returns the
// fingerprint of the server certificate, the original x509 error, and a
// nil-error.
func CheckPipelineEndpoint(endpoint string) (string, error, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", nil, emperror.Wrap(err, "failed to parse endpoint URL")
	}
	parsed.Path = path.Join(parsed.Path, "version")
	endpoint = parsed.String()

	var x509Err error
	response, err := http.Get(endpoint)
	if err != nil {
		x509Err = x509Error(err)
		if x509Err == nil {
			return "", nil, emperror.Wrap(err, "failed to connect to Pipeline")
		}

		insecure := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy:           http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
			},
		}
		response, err = insecure.Get(endpoint)
		if err != nil {
			return "", nil, emperror.Wrap(err, "failed to connect to Pipeline")
		}

	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", nil, errors.Errorf("failed to check Pipeline version: %s", response.Status)
	}

	version, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil, emperror.Wrap(err, "failed to check Pipeline version")
	}
	log.Debugf("Pipeline version: %q", string(version))

	if len(response.TLS.PeerCertificates) < 1 {
		return "", nil, errors.New("server certificate is missing")
	}
	hash, err := getFingerprint(response.TLS.PeerCertificates[0])
	if err != nil {
		return "", nil, err
	}
	return hex.EncodeToString(hash), x509Err, nil
}

func getFingerprint(cert *x509.Certificate) ([]byte, error) {
	switch cert.PublicKey.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
	default:
		return nil, errors.Errorf("server's certificate contains an unsupported type of public key: %T", cert.PublicKey)
	}

	hash := sha256.Sum256(cert.Raw)
	log.Debugf("server certificate Subject: %q, SHA256 hash: %x", cert.Subject, hash)
	return hash[0:], nil
}

func getServerCertFingerprint(certificates [][]byte) ([]byte, error) {
	if len(certificates) < 1 {
		return nil, errors.New("server certificate is missing")
	}
	cert, err := x509.ParseCertificate(certificates[0])
	if err != nil {
		return nil, emperror.Wrap(err, "failed to parse certificate from server")
	}

	return getFingerprint(cert)
}

func (c *banzaiCli) CloudinfoClient() *cloudinfo.APIClient {
	c.cloudinfoClientOnce.Do(func() {
		config := cloudinfo.NewConfiguration()
		config.BasePath = viper.GetString("cloudinfo.basepath")
		config.UserAgent = "banzai-cli/1.0.0/go"

		c.cloudinfoClient = cloudinfo.NewAPIClient(config)
	})

	return c.cloudinfoClient
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

func (c *banzaiCli) SetFingerprint(fingerprint string) {
	viper.Set("pipeline.tls-fingerprint", fingerprint)
	viper.Set("pipeline.tls-skip-verify", fingerprint != "")

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
