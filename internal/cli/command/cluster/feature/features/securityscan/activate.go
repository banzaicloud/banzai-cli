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

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type activateOptions struct {
	clustercontext.Context
	filePath string
}

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}
	ac := MakeFeatureActivator(banzaiCli)

	cmd := &cobra.Command{
		Use:           "activate",
		Aliases:       []string{"add", "enable", "install", "on"},
		Short:         fmt.Sprintf("Activate the %s feature of a cluster", featureName),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return ac.runActivate(options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("activate %s cluster feature for", featureName))

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

func (fa *featureActivator) runActivate(options activateOptions, args []string) error {

	if err := options.Init(args...); err != nil {
		return errors.Wrap(err, "failed to initialize options")
	}

	var req pipeline.ActivateClusterFeatureRequest
	if options.filePath == "" && fa.banzaiCLI.Interactive() {
		if err := fa.buildActivateReqInteractively(&req); err != nil {
			return errors.WrapIf(err, "failed to build activate request interactively")
		}
	}

	if err := fa.readActivateReqFromFileOrStdin(options.filePath, &req); err != nil {
		return errors.WrapIff(err, "failed to read %s cluster feature specification", featureName)
	}

	orgId := fa.banzaiCLI.Context().OrganizationID()
	clusterId := options.ClusterID()
	_, err := fa.banzaiCLI.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("activate %s cluster feature", featureName), err, req)
		log.Fatalf("could not activate %s cluster feature: %v", featureName, err)
	}

	log.Infof("feature %q started to activate", featureName)

	return nil
}

func (fa *featureActivator) readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (fa *featureActivator) buildActivateReqInteractively(req *pipeline.ActivateClusterFeatureRequest) error {

	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Edit the cluster feature activation request in your text editor?",
		},
		&edit,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if !edit {
		return fa.buildCustomAnchoreFeatureRequest(req)
	}

	spec, err := fa.securityScanSpecAsMap(nil)
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
		survey.WithValidator(fa.validateActivateClusterFeatureRequest),
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if err := json.Unmarshal([]byte(result), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

// featureActivator helper struct for gathering activate command realated operations
type featureActivator struct {
	specAssembler
}

// MakeFeatureActivator returns a reference to an featureActivator instance
func MakeFeatureActivator(banzaiCLI cli.Cli) *featureActivator {
	ac := new(featureActivator)
	ac.banzaiCLI = banzaiCLI
	return ac
}

func (fa *featureActivator) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
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

func (fa *featureActivator) buildCustomAnchoreFeatureRequest(activateRequest *pipeline.ActivateClusterFeatureRequest) error {

	anchoreConfig, err := fa.askForAnchoreConfig(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore configuration details")
	}

	policy, err := fa.askForPolicy(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	whiteLists, err := fa.askForWhiteLists()
	if err != nil {
		return errors.WrapIf(err, "failed to read whitelists")
	}

	webhookConfig, err := fa.askForWebHookConfig(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	securityScanFeatureRequest := new(SecurityScanFeatureSpec)
	securityScanFeatureRequest.CustomAnchore = *anchoreConfig
	securityScanFeatureRequest.Policy = *policy
	securityScanFeatureRequest.ReleaseWhiteList = whiteLists
	securityScanFeatureRequest.WebhookConfig = *webhookConfig

	ssfMap, err := fa.securityScanSpecAsMap(securityScanFeatureRequest)
	if err != nil {
		return errors.WrapIf(err, "failed to transform request to map")
	}

	activateRequest.Spec = ssfMap

	return nil
}

func (fa *featureActivator) askForWhiteLists() ([]releaseSpec, error) {

	addMore := true
	releaseWhiteList := make([]releaseSpec, 0)

	for addMore {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Add a release whitelist item to the security scan? ",
			},
			&addMore,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}

		if !addMore {
			continue
		}

		item, err := fa.askForWhiteListItem()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to read release whitelist item")
		}
		releaseWhiteList = append(releaseWhiteList, *item)
	}

	return releaseWhiteList, nil
}

func (fa *featureActivator) askForWhiteListItem() (*releaseSpec, error) {

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

func (fa *featureActivator) validateActivateClusterFeatureRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}
