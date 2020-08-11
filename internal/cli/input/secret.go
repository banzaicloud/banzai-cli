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

package input

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	// "gopkg.in/yaml.v2" -- could not be used for kubernetes types
	"github.com/ghodss/yaml"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

func AskSecret(banzaiCli cli.Cli, orgID int32, cloud string) (string, error) {
	var secretName string

	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{
		Type_: optional.NewString(cloud),
	})
	if err != nil {
		return "", errors.WrapIf(utils.ConvertError(err), "could not get secret")
	}

	secretOptions := make([]string, len(secrets))
	secretIds := make(map[string]string, len(secrets))
	for i, s := range secrets {
		secretOptions[i] = s.Name
		secretIds[s.Name] = s.Id
	}
	err = survey.AskOne(&survey.Select{Message: "Secret:", Options: secretOptions}, &secretName, survey.WithValidator(survey.Required))
	if err != nil {
		return "", errors.WrapIf(err, "failed to select secret")
	}

	return secretIds[secretName], nil
}

const AwsRegionKey = "AWS_DEFAULT_REGION"

// GetAmazonCredentialsRegion extracts the local credentials from env vars and user profile while ensuring a region
func GetAmazonCredentialsRegion(profile string, defaultRegion string, assumeRole string) (string, string, map[string]string, error) {
	id, out, err := GetAmazonCredentials(profile, assumeRole)
	if err != nil {
		return id, "", out, err
	}

	if out[AwsRegionKey] == "" {
		if defaultRegion == "" {
			return "", "", nil, errors.New("no default AWS region is set")
		}
		out[AwsRegionKey] = defaultRegion
	}
	return id, out[AwsRegionKey], out, err
}

// GetAmazonCredentials extracts the local credentials from env vars and user profile
func GetAmazonCredentials(profile string, assumeRole string) (string, map[string]string, error) {
	/* create a new session, which is basically the same as the following, but may also contain a region
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
		}) */
	var sess *session.Session
	var err error

	// use direct auth based on exported env vars
	if id, ok := os.LookupEnv("AWS_ACCESS_KEY_ID"); ok {
		if profile != "" {
			log.Warnf("AWS profile `%s` is overridden to `%s` by AWS_ACCESS_KEY_ID env var explicitly", profile, id)
		}
		sess, err = session.NewSession(&aws.Config{})
	} else {
		sess, err = session.NewSessionWithOptions(session.Options{
			Profile: profile,
		})
	}

	if err != nil {
		return "", nil, err
	}

	var creds credentials.Value

	if assumeRole != "" {
		options := func(p *stscreds.AssumeRoleProvider) {
			p.Duration = 1 * time.Hour
		}
		creds, err = stscreds.NewCredentials(sess, assumeRole, options).Get()
		if err != nil {
			return "", nil, err
		}
	} else {
		creds, err = sess.Config.Credentials.Get()
		if err != nil {
			return "", nil, err
		}
	}

	out := map[string]string{
		"AWS_ACCESS_KEY_ID":     creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":     creds.SessionToken,
	}

	if sess.Config.Region != nil {
		out[AwsRegionKey] = *sess.Config.Region
	}

	return creds.AccessKeyID, out, nil
}

// GetCurrentKubecontext extracts the Kubernetes context selected locally
func GetCurrentKubecontext() (string, []byte, error) {
	c := exec.Command("kubectl", "config", "view", "--minify", "--raw")
	out, err := c.Output()
	if err != nil {
		return "", nil, errors.WrapIf(err, "failed to query current context from kubectl")
	}

	var parsed v1.Config
	if err := yaml.Unmarshal(out, &parsed); err != nil {
		return "", nil, errors.WrapIf(err, "failed to parse local configuration")
	}

	if len(parsed.AuthInfos) != 1 {
		return "", nil, errors.New("kubernetes config doesn't contain a single user definition")
	}
	authConf := parsed.AuthInfos[0].AuthInfo.AuthProvider

	if authConf != nil && authConf.Config["cmd-path"] != "" {
		// TODO add support
		return "", nil, errors.Errorf("kubernetes authorization helpers (%s) are not supported", filepath.Base(authConf.Config["cmd-path"]))
	}

	return parsed.CurrentContext, out, nil
}

// RewriteLocalhostToHostDockerInternal rewrites a minified kubeconfig to use `host.docker.internal` in place of localhost
func RewriteLocalhostToHostDockerInternal(in []byte) ([]byte, error) {
	var parsed v1.Config
	if err := yaml.Unmarshal(in, &parsed); err != nil {
		return nil, errors.WrapIf(err, "failed to parse local configuration")
	}
	if len(parsed.Clusters) != 1 {
		return nil, errors.New("kubernetes config doesn't contain a single cluster definition")
	}
	serverURL, err := url.Parse(parsed.Clusters[0].Cluster.Server)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to parse server URL")
	}

	if host := serverURL.Hostname(); host == "localhost" || host == "127.0.0.1" {
		port := serverURL.Port()
		serverURL.Host = "host.docker.internal"
		if port != "" {
			serverURL.Host += ":" + port
		}
		parsed.Clusters[0].Cluster.Server = serverURL.String()
		if len(parsed.Clusters[0].Cluster.CertificateAuthorityData) > 0 {
			// kubernetes has no way to disable cert common name check only
			parsed.Clusters[0].Cluster.CertificateAuthorityData = []byte{}
			parsed.Clusters[0].Cluster.InsecureSkipTLSVerify = true
		}
	}

	out, err := yaml.Marshal(parsed)
	return out, errors.WrapIf(err, "failed to parse local configuration")
}
