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
	"encoding/json"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type activateManager struct {
	baseManager
	specAssembler
}

func NewActivateManager() features.ActivateManager {
	return new(activateManager)
}

func (am *activateManager) BuildRequestInteractively(cli.Cli, clustercontext.Context) (*pipeline.ActivateClusterFeatureRequest, error) {
	var req pipeline.ActivateClusterFeatureRequest

	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Edit the cluster feature activation request in your text editor?",
		},
		&edit,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !edit {
		if err := am.buildAnchoreConfigSpec(&req); err != nil {
			return nil, err
		}
		return &req, nil
	}

	spec, err := am.securityScanSpecAsMap(nil)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to decode spec into map")
	}

	req.Spec = spec

	content, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, errors.WrapIf(err, "failed to marshal request to JSON")
	}
	var result string
	if err := survey.AskOne(
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true,
		},
		&result,
		survey.WithValidator(am.ValidateRequest),
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if err := json.Unmarshal([]byte(result), &req); err != nil {
		return nil, errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return &req, nil
}

func (a activateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func (am *activateManager) readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (am *activateManager) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
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

func (am *activateManager) buildAnchoreConfigSpec(activateRequest *pipeline.ActivateClusterFeatureRequest) error {

	anchoreConfig, err := am.askForAnchoreConfig(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore configuration details")
	}

	policy, err := am.askForPolicy(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	whiteLists, err := am.askForWhiteLists()
	if err != nil {
		return errors.WrapIf(err, "failed to read whitelists")
	}

	webhookConfig, err := am.askForWebHookConfig(nil)
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	securityScanFeatureRequest := new(SecurityScanFeatureSpec)
	securityScanFeatureRequest.CustomAnchore = *anchoreConfig
	securityScanFeatureRequest.Policy = *policy
	securityScanFeatureRequest.ReleaseWhiteList = whiteLists
	securityScanFeatureRequest.WebhookConfig = *webhookConfig

	ssfMap, err := am.securityScanSpecAsMap(securityScanFeatureRequest)
	if err != nil {
		return errors.WrapIf(err, "failed to transform request to map")
	}

	activateRequest.Spec = ssfMap

	return nil
}
