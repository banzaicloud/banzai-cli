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

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}

	cmd := &cobra.Command{
		Use:           "activate",
		Aliases:       []string{"add", "enable", "install", "on"},
		Short:         fmt.Sprintf("Activate the %s feature of a cluster", featureName),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runActivate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("activate %s cluster feature for", featureName))

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type activateOptions struct {
	clustercontext.Context
	filePath string
}

func runActivate(banzaiCli cli.Cli, options activateOptions, args []string) error {

	if err := options.Init(args...); err != nil {
		return errors.Wrap(err, "failed to initialize options")
	}

	var req pipeline.ActivateClusterFeatureRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if err := buildActivateReqInteractively(banzaiCli, &req); err != nil {
			return errors.WrapIf(err, "failed to build activate request interactively")
		}
	} else {
		if err := readActivateReqFromFileOrStdin(options.filePath, &req); err != nil {
			return errors.WrapIff(err, "failed to read %s cluster feature specification", featureName)
		}
	}

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()
	_, err := banzaiCli.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("activate %s cluster feature", featureName), err, req)
		log.Fatalf("could not activate %s cluster feature: %v", featureName, err)
	}

	log.Infof("feature %q started to activate", featureName)

	return nil
}

func readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func buildActivateReqInteractively(
	banzaiCli cli.Cli,
	req *pipeline.ActivateClusterFeatureRequest,
) error {

	aCommander := MakeActivateCommander(banzaiCli)

	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Do you want to edit the cluster feature activation request in your text editor?",
		},
		&edit,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if !edit {
		return aCommander.buildCustomAnchoreFeatureRequest(req)
	}

	spec, err := aCommander.securityScanSpecAsMap(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to decode spec into map")
	}

	req.Spec = spec

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}
	var result string
	if err := survey.AskOne(
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true,
		},
		&result,
		survey.WithValidator(validateActivateClusterFeatureRequest),
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if err := json.Unmarshal([]byte(result), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func validateActivateClusterFeatureRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func buildScurityscanFeatureRequest(banzaiCli cli.Cli, req *pipeline.ActivateClusterFeatureRequest) error {

	return nil
}

// activateCommander helper struct for gathering activate command realated operations
type activateCommander struct {
	banzaiCLI cli.Cli
}

// MakeActivateCommander returns a reference to an activateCommander instance
func MakeActivateCommander(banzaiCLI cli.Cli) *activateCommander {
	ac := new(activateCommander)
	ac.banzaiCLI = banzaiCLI
	return ac
}

func (ac *activateCommander) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
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

func (ac *activateCommander) askForAnchoreConfig() (*anchoreSpec, error) {

	var customAnchore bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Would you like to configure a custom anchore instance? ",
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
		},
		&anchoreURL,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read custom Anchore URL")
	}

	secretID, err := ac.askForSecret()
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read secret for accessing custom Anchore ")
	}

	return &anchoreSpec{
		Enabled:  true,
		Url:      anchoreURL,
		SecretID: secretID,
	}, nil
}

func (ac *activateCommander) askForSecret() (string, error) {
	const (
		PasswordSecretType = "password"
	)

	orgID := ac.banzaiCLI.Context().OrganizationID()

	secrets, _, err := ac.banzaiCLI.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(PasswordSecretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there are no secrets with type %q", PasswordSecretType))
	}

	options := make([]string, len(secrets))
	for i, s := range secrets {
		options[i] = s.Name
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a secret to access the custom Anchore instance:",
			Options: options,
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

func (ac *activateCommander) buildCustomAnchoreFeatureRequest(activateRequest *pipeline.ActivateClusterFeatureRequest) error {

	anchoreConfig, err := ac.askForAnchoreConfig()
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore configuration details")
	}

	policy, err := ac.askForPolicy()
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	whiteLists, err := ac.askForWhiteLists()
	if err != nil {
		return errors.WrapIf(err, "failed to read whitelists")
	}

	webhookConfig, err := ac.askForWebHookConfig()
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	securityScanFeatureRequest := new(SecurityScanFeatureSpec)
	securityScanFeatureRequest.CustomAnchore = *anchoreConfig
	securityScanFeatureRequest.Policy = *policy
	securityScanFeatureRequest.ReleaseWhiteList = whiteLists
	securityScanFeatureRequest.WebhookConfig = *webhookConfig

	ssfMap, err := ac.securityScanSpecAsMap(securityScanFeatureRequest)

	activateRequest.Spec = ssfMap

	return nil
}

func (ac *activateCommander) askForPolicy() (*policySpec, error) {

	type policy struct {
		name string
		id   string
	}
	// todo add all supported policies here
	policies := []policy{{"Default bundle", "1"}}

	options := make([]string, len(policies))
	for i, s := range policies {
		options[i] = s.name
	}

	var policyName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a policy for the Anchor Engine:",
			Options: options,
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

func (ac *activateCommander) askForWhiteLists() ([]releaseSpec, error) {

	addMore := true
	releaseWhiteList := make([]releaseSpec, 0)

	for addMore {

		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Would you like to add a release whitelist item to the security scan? ",
			},
			&addMore,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}

		if !addMore {
			continue
		}

		item, err := ac.askForWhiteListItem()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to read release whitelist item")
		}
		releaseWhiteList = append(releaseWhiteList, *item)

	}

	return releaseWhiteList, nil
}

func (ac *activateCommander) askForWebHookConfig() (*webHookConfigSpec, error) {
	var enable bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Would you like to configure the security scan webhook?",
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
		},
		&selector,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select policy")
	}

	var namespaces string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the comma separated list of namespaces",
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

func (ac *activateCommander) askForWhiteListItem() (*releaseSpec, error) {

	var releaseName string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the name of the release whitelist item:",
		},
		&releaseName,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the name of the release whitelist item")
	}

	var reason string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the reason of the release whitelist item:",
		},
		&reason,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the reason of the release whitelist item")
	}

	var regexp string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the regexp for the release whitelist item:",
		},
		&regexp,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the regexp of the release whitelist item")
	}

	return &releaseSpec{
		Name:   releaseName,
		Reason: reason,
		Regexp: regexp,
	}, nil
}
