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
	"time"

	"github.com/antihax/optional"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type createOptions struct {
	file     string
	wait     bool
	interval int
}

// NewCreateCommand creates a new cobra.Command for `banzai cluster create`.
func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:          "create",
		Aliases:      []string{"c"},
		Short:        "Create a cluster",
		Long:         "Create cluster based on json stdin or interactive session",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Cluster descriptor file")
	flags.BoolVarP(&options.wait, "wait", "w", false, "Wait for cluster creation")
	flags.IntVarP(&options.interval, "interval", "i", 10, "Interval in seconds for polling cluster status")

	return cmd
}

func runCreate(banzaiCli cli.Cli, options createOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	out := map[string]interface{}{}

	if banzaiCli.Interactive() {
		err := buildInteractiveCreateRequest(banzaiCli, options, orgID, out)
		if err != nil {
			return err
		}
	} else { // non-interactive
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return emperror.WrapWith(err, "failed to read", "filename", filename)
		}

		log.Debugf("%d bytes read", len(raw))

		if err := validateClusterCreateRequest(raw); err != nil {
			return emperror.Wrap(err, "failed to parse create cluster request")
		}

		if err := utils.Unmarshal(raw, &out); err != nil {
			return emperror.Wrap(err, "failed to unmarshal create cluster request")
		}
	}

	log.Debugf("create request: %#v", out)
	cluster, _, err := banzaiCli.Client().ClustersApi.CreateCluster(context.Background(), orgID, out)
	if err != nil {
		cli.LogAPIError("create cluster", err, out)
		return emperror.Wrap(err, "failed to create cluster")
	}

	log.Info("cluster is being created")
	if options.wait {
		for {
			cluster, _, err := banzaiCli.Client().ClustersApi.GetCluster(context.Background(), orgID, cluster.Id)
			if err != nil {
				cli.LogAPIError("create cluster", err, out)
			} else {
				format.ClusterShortWrite(banzaiCli, cluster)
				if cluster.Status != "CREATING" {
					return nil
				}

				time.Sleep(time.Duration(options.interval) * time.Second)
			}
		}
	} else {
		log.Infof("you can check its status with the command `banzai cluster get %q`", out["name"])
		format.ClusterShortWrite(banzaiCli, cluster)
	}
	return nil
}

func validateClusterCreateRequest(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		if bytes, ok := val.([]byte); ok {
			str = string(bytes)
		} else {
			return errors.New("value is not a string or []byte")
		}
	}

	decoder := json.NewDecoder(strings.NewReader(str))

	var typer struct{ Type string }
	err := decoder.Decode(&typer)
	if err != nil {
		return emperror.Wrap(err, "invalid JSON request")
	}

	decoder = json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()

	if typer.Type == "" {
		err = decoder.Decode(&pipeline.CreateClusterRequest{})
	} else {
		err = decoder.Decode(&pipeline.CreateClusterRequestV2{})
	}
	return emperror.Wrap(err, "invalid request")
}

func buildInteractiveCreateRequest(banzaiCli cli.Cli, options createOptions, orgID int32, out map[string]interface{}) error {
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
			if err := utils.Unmarshal(raw, &out); err != nil {
				return emperror.Wrap(err, "failed to parse CreateClusterRequest")
			}

			break
		}
	}

	if out["cloud"] == nil && out["type"] == nil {
		err := buildDefaultRequest(out)
		if err != nil {
			return err
		}
	}

	cloud, ok := out["cloud"].(string)
	if !ok || cloud == "" {
		Type, _ := out["type"].(string)
		switch Type {
		case "pke-on-azure":
			cloud = "azure"
		default:
			return errors.New("couldn't determine cloud provider from request")
		}
	}

	secretID, err := buildSecretChoice(banzaiCli, orgID, cloud, out)
	if err != nil {
		return err
	}

	if out["name"] == nil || out["name"] == "" {
		name := fmt.Sprintf("%s%d", os.Getenv("USER"), os.Getpid())
		_ = survey.AskOne(&survey.Input{Message: "Cluster name:", Default: name}, &name, nil)
		out["name"] = name
	}

	if out["type"] == "pke-on-azure" && out["resourceGroup"] == "" {
		rgs, _, err := banzaiCli.Client().InfoApi.GetResourceGroups(context.Background(), orgID, secretID)
		if err != nil {
			return emperror.Wrap(err, "can't list resource groups")
		}

		var rg string
		if err = survey.AskOne(&survey.Select{Message: "Resource group:", Options: rgs}, &rg, nil); err == nil {
			out["resourceGroup"] = rg
		} else {
			log.Error("no resource group selected")
		}
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

		_ = survey.AskOne(&survey.Editor{Message: "Create cluster request:", Default: content, HideDefault: true, AppendDefault: true}, &content, validateClusterCreateRequest)
		if err := json.Unmarshal([]byte(content), &out); err != nil {
			log.Errorf("can't parse request: %v", err)
		}
	}

	var create bool
	_ = survey.AskOne(
		&survey.Confirm{
			Message: fmt.Sprintf("Do you want to CREATE the cluster %q now?", out["name"]),
		},
		&create,
		nil,
	)

	if !create {
		return errors.New("cluster creation cancelled")
	}

	return nil
}
func getProviders() map[string]interface{} {
	return map[string]interface{}{
		"pke-on-aws": pipeline.CreateClusterRequest{
			Cloud:    "amazon",
			Location: "us-east-2",
			Properties: map[string]interface{}{
				"pke": pipeline.CreatePkeProperties{
					NodePools: []pipeline.NodePoolsPke{
						{
							Name:     "master",
							Roles:    []string{"master", "worker"},
							Provider: "amazon",
							ProviderConfig: map[string]interface{}{
								"autoScalingGroup": map[string]interface{}{
									"instanceType": "c5.large",
									"zones":        []string{"us-east-2a"},
									"spotPrice":    "",
									"size": map[string]interface{}{
										"desired": 1,
										"min":     1,
										"max":     1,
									},
								},
							},
						},
					},
					Kubernetes: pipeline.CreatePkePropertiesKubernetes{
						Version: "v1.14.2",
						Rbac: pipeline.CreatePkePropertiesKubernetesRbac{
							Enabled: true,
						},
					},
					Cri: pipeline.CreatePkePropertiesCri{
						Runtime: "containerd",
					},
				},
			},
		},
		"pke-on-azure": pipeline.CreatePkeOnAzureClusterRequest{
			Type:     "pke-on-azure",
			Location: "westus2",
			Nodepools: []pipeline.PkeOnAzureNodePool{
				{
					Name:         "master",
					Roles:        []string{"master", "worker"},
					Autoscaling:  false,
					MinCount:     1,
					MaxCount:     1,
					Count:        1,
					InstanceType: "Standard_D2s_v3",
				},
			},
			Kubernetes: pipeline.CreatePkeClusterKubernetes{
				Version: "1.14.2",
				Rbac:    true,
			},
		},
		"ack": pipeline.CreateClusterRequest{
			Cloud: "alibaba",
			Properties: map[string]interface{}{
				"ack": pipeline.CreateAckPropertiesAck{},
			},
		},
		"aks": pipeline.CreateClusterRequest{
			Cloud:    "azure",
			Location: "westus2",
			Properties: map[string]interface{}{
				"aks": pipeline.CreateAksPropertiesAks{},
			},
		},
		"eks": pipeline.CreateClusterRequest{
			Cloud:    "amazon",
			Location: "us-east-2",
			Properties: map[string]interface{}{
				"eks": pipeline.CreateEksPropertiesEks{},
			},
		},
		"gke": pipeline.CreateClusterRequest{
			Cloud: "google",
			Properties: map[string]interface{}{
				"gke": pipeline.CreateGkePropertiesGke{},
			},
		},
		"oke": pipeline.CreateClusterRequest{
			Cloud: "oracle",
			Properties: map[string]interface{}{
				"oke": pipeline.CreateUpdateOkePropertiesOke{},
			},
		},
	}
}

func buildDefaultRequest(out map[string]interface{}) error {
	providers := getProviders()
	providerNames := make([]string, 0, len(providers))

	for provider := range providers {
		providerNames = append(providerNames, provider)
	}

	var providerName string

	_ = survey.AskOne(&survey.Select{Message: "Provider:", Help: "Select the provider to use", Options: providerNames}, &providerName, nil)

	if provider, ok := providers[providerName]; ok {
		marshalled, err := json.Marshal(provider)
		if err != nil {
			return emperror.Wrap(err, "failed to marshal request template")
		}

		utils.Unmarshal(marshalled, &out)
	}
	return nil
}

func buildSecretChoice(banzaiCli cli.Cli, orgID int32, cloud string, out map[string]interface{}) (string, error) {

	if id, ok := out["secretId"].(string); id != "" && ok {
		return id, nil
	}

	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(cloud)})
	if err != nil {
		return "", emperror.Wrap(err, "could not list secrets")
	}

	// get ID from Name + validate
	if name, ok := out["secretName"].(string); name != "" && ok {

		for _, secret := range secrets {
			if secret.Name == name {
				return secret.Id, nil
			}
		}
		return "", errors.New(fmt.Sprintf("can't find %s secret %q", cloud, name))
	}

	// offer secret choices

	secretNames := make([]string, len(secrets))
	secretIDs := make(map[string]string)

	for i, secret := range secrets {
		secretNames[i] = secret.Name
		secretIDs[secret.Name] = secret.Id
	}

	var name string
	if err = survey.AskOne(&survey.Select{Message: "Secret:", Help: "Select the secret to use for creating cloud resources", Options: secretNames}, &name, nil); err != nil {
		return "", emperror.Wrap(err, "no secret set")
	}
	out["secretName"] = name
	return secretIDs[name], nil
}
