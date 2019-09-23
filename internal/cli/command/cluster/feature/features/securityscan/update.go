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

package securityscan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"change", "modify", "set"},
		Short:   fmt.Sprintf("Update the %s feature of a cluster", featureName),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpdate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("update the %s cluster feature for", featureName))

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type updateOptions struct {
	clustercontext.Context
	filePath string
}

// featureUpdater struct for gathering helper operations for the feature update
type featureUpdater struct {
	banzaiCLI cli.Cli
}

func MakeFeatureUpdater(banzaiCLI cli.Cli) *featureUpdater {
	fu := new(featureUpdater)
	fu.banzaiCLI = banzaiCLI
	return fu
}

func (fu *featureUpdater) getSecurityScanFeature(orgID int32, clusterID int32) (map[string]interface{}, error) {

	clusterFeatureDetails, _, err := fu.banzaiCLI.Client().ClusterFeaturesApi.ClusterFeatureDetails(context.Background(), orgID, clusterID, featureName)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to retrieve the feature to update")
	}

	//var securityFeatureSpec *SecurityScanFeatureSpec
	//if err := mapstructure.Decode(clusterFeatureDetails.Spec, securityFeatureSpec); err != nil {
	//	return nil, errors.WrapIf(err, "failed to decode the feature to update")
	//}

	return clusterFeatureDetails.Spec, nil
}

func ValidateUpdateClusterFeatureRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func runUpdate(banzaiCli cli.Cli, options updateOptions, args []string) error {
	if err := options.Init(args...); err != nil {
		return errors.Wrap(err, "failed to initialize options")
	}

	fu := MakeFeatureUpdater(banzaiCli)

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()

	securityScanFeatureSpec, err := fu.getSecurityScanFeature(orgId, clusterId)
	if err != nil {
		return errors.WrapIf(err, "failed to update feature")
	}

	req := new(pipeline.UpdateClusterFeatureRequest)
	req.Spec = securityScanFeatureSpec

	if options.filePath == "" && banzaiCli.Interactive() {
		if err := fu.buildUpdateReqInteractively(options, req); err != nil {
			return errors.WrapIf(err, "failed to build update request interactively")
		}
	} else {
		if err := readUpdateReqFromFileOrStdin(options.filePath, req); err != nil {
			return errors.WrapIff(err, "failed to read %s cluster feature specification", featureName)
		}
	}

	resp, err := banzaiCli.Client().ClusterFeaturesApi.UpdateClusterFeature(context.Background(), orgId, clusterId, featureName, *req)
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("activate %s cluster feature", featureName), err, resp.Request)
		log.Fatalf("could not activate %s cluster feature: %v", featureName, err)
	}

	log.Infof("feature %q started to update", featureName)

	return nil
}

func readUpdateReqFromFileOrStdin(filePath string, req *pipeline.UpdateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (fu *featureUpdater) buildUpdateReqInteractively(_ updateOptions, req *pipeline.UpdateClusterFeatureRequest) error {
	var edit bool
	if err := survey.AskOne(&survey.Confirm{Message: "Edit the cluster feature update request in your text editor?"}, &edit); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if !edit {
		return fu.buildCustomAnchoreFeatureRequest(req)
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}

	prompt := &survey.Editor{Default: string(content), HideDefault: true, AppendDefault: true}

	var updated string
	if err := survey.AskOne(prompt, &updated); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if err := json.Unmarshal([]byte(updated), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func (fu *featureUpdater) buildCustomAnchoreFeatureRequest(updateRequest *pipeline.UpdateClusterFeatureRequest) error {

	// get the type from the req
	securityFeatureSpec := new(SecurityScanFeatureSpec)
	if err := mapstructure.Decode(updateRequest.Spec, securityFeatureSpec); err != nil {
		return errors.WrapIf(err, "failed to decode the feature to update")
	}

	anchoreConfig, err := fu.askForAnchoreConfig(securityFeatureSpec.CustomAnchore)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore configuration details")
	}

	policy, err := fu.askForPolicy(securityFeatureSpec.Policy)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	// todo whitelist updates not supported fro now
	webhookConfig, err := fu.askForWebHookConfig(securityFeatureSpec.WebhookConfig)
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	securityScanFeatureRequest := new(SecurityScanFeatureSpec)
	securityScanFeatureRequest.CustomAnchore = *anchoreConfig
	securityScanFeatureRequest.Policy = *policy
	securityScanFeatureRequest.ReleaseWhiteList = securityFeatureSpec.ReleaseWhiteList
	securityScanFeatureRequest.WebhookConfig = *webhookConfig

	ssfMap, err := fu.securityScanSpecAsMap(securityScanFeatureRequest)

	updateRequest.Spec = ssfMap

	return nil
}

func (fu *featureUpdater) askForAnchoreConfig(currentSpec anchoreSpec) (*anchoreSpec, error) {

	var customAnchore bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Configure a custom anchore instance? ",
		},
		&customAnchore,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !customAnchore {
		return &anchoreSpec{
			Enabled: false,
		}, nil
	}

	// custom anchore config
	var anchoreURL string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the custom anchore URL:",
			Default: currentSpec.Url,
		},
		&anchoreURL,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read custom Anchore URL")
	}

	secretID, err := fu.askForSecret(currentSpec.SecretID)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read secret for accessing custom Anchore ")
	}

	return &anchoreSpec{
		Enabled:  true,
		Url:      anchoreURL,
		SecretID: secretID,
	}, nil
}

func (fu *featureUpdater) askForSecret(secretID string) (string, error) {
	const (
		PasswordSecretType = "password"
	)

	orgID := fu.banzaiCLI.Context().OrganizationID()

	secrets, _, err := fu.banzaiCLI.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(PasswordSecretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there are no secrets with type %q", PasswordSecretType))
	}

	options := make([]string, len(secrets))
	currentSecretName := ""
	for i, s := range secrets {
		options[i] = s.Name
		if s.Id == secretID {
			currentSecretName = s.Name
		}
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a secret to access the custom Anchore instance:",
			Options: options,
			Default: currentSecretName,
		},
		&secretName,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve secrets")
	}

	for _, s := range secrets {
		if s.Name == secretName {
			return s.Id, nil
		}
	}

	return "", errors.Errorf("no secret with name %q", secretName)
}

func (fu *featureUpdater) askForPolicy(currentPolicy policySpec) (*policySpec, error) {
	type policy struct {
		name string
		id   string
	}
	// todo add all supported policies here
	policies := []policy{{"Default bundle", "1"}}

	options := make([]string, len(policies))
	currentPolicyName := ""
	for i, s := range policies {
		options[i] = s.name
		if s.id == currentPolicy.PolicyID {
			currentPolicyName = s.name
		}
	}

	var policyName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a policy for the Anchor Engine:",
			Options: options,
			Default: currentPolicyName,
		},
		&policyName,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select policy")
	}

	var selected policy
	for _, s := range policies {
		if s.name == policyName {
			selected = s
		}
	}

	return &policySpec{
		PolicyID: selected.id,
	}, nil
}

func (fu *featureUpdater) askForWebHookConfig(currentWH webHookConfigSpec) (*webHookConfigSpec, error) {
	var enable bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Configure the security scan webhook?",
		},
		&enable,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !enable {
		return &webHookConfigSpec{
			Enabled: false,
		}, nil
	}

	var selector string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please choose the selector for the webhook configuration:",
			Options: []string{"Exclude", "Include"},
			Default: currentWH.Selector,
		},
		&selector,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select policy")
	}

	var namespaces string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the comma separated list of namespaces:",
			Default: strings.Join(currentWH.Namespaces, ","),
		},
		&namespaces,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read namespaces")
	}

	return &webHookConfigSpec{
		Enabled:    true,
		Selector:   selector,
		Namespaces: strings.Split(namespaces, ","),
	}, nil
}

func (fu *featureUpdater) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
	// fill the structure of the config - make filling up the values easier
	if spec == nil {
		spec = &SecurityScanFeatureSpec{
			CustomAnchore:    anchoreSpec{},
			Policy:           policySpec{},
			ReleaseWhiteList: nil,
			WebhookConfig:    webHookConfigSpec{},
		}
	}

	var specMap map[string]interface{}
	if err := mapstructure.Decode(spec, &specMap); err != nil {
		return nil, err
	}

	return specMap, nil
}
