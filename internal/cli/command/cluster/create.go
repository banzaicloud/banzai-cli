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

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type createOptions struct {
	file string
}

// NewCreateCommand creates a new cobra.Command for `banzai cluster create`.
func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a cluster",
		Long:    "Create cluster based on json stdin or interactive session",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runCreate(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Cluster descriptor file")

	return cmd
}

func runCreate(banzaiCli cli.Cli, options createOptions) {
	orgID := input.GetOrganization(banzaiCli)

	out := pipeline.CreateClusterRequest{}

	if isInteractive() {
		var content string
		var fileName = options.file

		for {
			if fileName == "" {
				_ = survey.AskOne(
					&survey.Input{
						Message: "Load a JSON or YAML file:",
						Default: "skip",
						Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML Cluster creation request. Leave empty to cancel.",
					},
					&fileName,
					nil,
				)
				if fileName == "skip" || fileName == "" {
					break
				}
			}

			if raw, err := ioutil.ReadFile(fileName); err != nil {
				fileName = "" // reset fileName so that we can ask for one

				log.Errorf("failed to read file %q: %v", fileName, err)

				continue
			} else {
				if err := unmarshal(raw, &out); err != nil {
					log.Fatalf("failed to parse CreateClusterRequest: %v", err)
				}

				break
			}
		}

		if out.Properties == nil || len(out.Properties) == 0 {
			providers := map[string]struct {
				cloud    string
				property interface{}
			}{
				"acsk": {cloud: "alibaba", property: new(pipeline.CreateAckPropertiesAcsk)},
				"aks":  {cloud: "azure", property: new(pipeline.CreateAksPropertiesAks)},
				"eks":  {cloud: "amazon", property: new(pipeline.CreateEksPropertiesEks)},
				"gke":  {cloud: "google", property: new(pipeline.CreateEksPropertiesEks)},
				"oke":  {cloud: "oracle", property: map[string]interface{}{}},
			}

			providerNames := make([]string, 0, len(providers))

			for provider := range providers {
				providerNames = append(providerNames, provider)
			}

			var providerName string

			_ = survey.AskOne(&survey.Select{Message: "Provider:", Help: "Select the provider to use", Options: providerNames}, &providerName, nil)

			if provider, ok := providers[providerName]; ok {
				out.Properties = map[string]interface{}{providerName: provider.property}
				out.Cloud = provider.cloud
			}
		}
		if out.SecretId == "" && out.SecretName == "" {
			secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(out.Cloud)})

			if err != nil {
				log.Errorf("could not list secrets: %v", err)
			} else {
				secretNames := make([]string, len(secrets))

				for i, secret := range secrets {
					secretNames[i] = secret.Name
				}

				_ = survey.AskOne(&survey.Select{Message: "Secret:", Help: "Select the secret to use for creating cloud resources", Options: secretNames}, &out.SecretName, nil)
			}
		}

		if out.Name == "" {
			name := fmt.Sprintf("%s%s%d", os.Getenv("USER"), out.Cloud, os.Getpid())
			_ = survey.AskOne(&survey.Input{Message: "Cluster name:", Default: name}, &out.Name, nil)
		}

		for {
			if bytes, err := json.MarshalIndent(out, "", "  "); err != nil {
				log.Errorf("failed to marshal request: %v", err)
				log.Debugf("Request: %#v", out)
			} else {
				content = string(bytes)
				_, _ = fmt.Fprintf(os.Stderr, "The current state of the request:\n\n%s\n", content)
			}

			var open bool
			_ = survey.AskOne(&survey.Confirm{Message: "Do you want to edit the cluster request in your text editor?"}, &open, nil)
			if !open {
				break
			}

			///fmt.Printf("BEFORE>>>\n%v<<<\n", content)
			_ = survey.AskOne(&survey.Editor{Message: "Create cluster request:", Default: content, HideDefault: true, AppendDefault: true}, &content, validateClusterCreateRequest)
			///fmt.Printf("AFTER>>>\n%v<<<\n", content)
			if err := json.Unmarshal([]byte(content), &out); err != nil {
				log.Errorf("can't parse request: %v", err)
			}
		}

		var create bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: fmt.Sprintf("Do you want to CREATE the cluster %q now?", out.Name),
			},
			&create,
			nil,
		)

		if !create {
			log.Fatal("cluster creation cancelled")
		}
	} else { // non-interactive
		var raw []byte
		var err error
		filename := options.file

		if filename != "" {
			raw, err = ioutil.ReadFile(filename)
		} else {
			raw, err = ioutil.ReadAll(os.Stdin)
			filename = "stdin"
		}

		if err != nil {
			log.Fatalf("failed to read %s: %v", filename, err)
		}

		if err := unmarshal(raw, &out); err != nil {
			log.Fatalf("failed to parse CreateClusterRequest: %v", err)
		}
	}

	log.Debugf("create request: %#v", out)
	cluster, _, err := banzaiCli.Client().ClustersApi.CreateCluster(context.Background(), orgID, out)
	if err != nil {
		cli.LogAPIError("create cluster", err, out)
		log.Fatalf("failed to create cluster: %v", err)
	}

	log.Info("cluster is being created")
	log.Infof("you can check its status with the command `banzai cluster get %q`", out.Name)
	Out1(cluster, []string{"Id", "Name"})
}

func validateClusterCreateRequest(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("value is not a string")
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()

	return emperror.Wrap(decoder.Decode(&pipeline.CreateClusterRequest{}), "not a valid JSON request")
}
