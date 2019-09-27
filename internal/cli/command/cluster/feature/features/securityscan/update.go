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

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/mitchellh/mapstructure"
)

type updateManager struct {
	baseManager
	specAssembler
}

func NewUpdateManager() features.UpdateManager {
	return new(updateManager)
}

func (u *updateManager) ValidateRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func (u *updateManager) BuildRequestInteractively(cli cli.Cli, req *pipeline.UpdateClusterFeatureRequest) error {
	var edit bool
	if err := survey.AskOne(&survey.Confirm{Message: "Edit the cluster feature update request in your text editor?"},
		&edit); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if !edit {
		return u.buildCustomAnchoreFeatureRequest(req)
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}

	var result string
	if err := survey.AskOne(
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true},
		&result,
		survey.WithValidator(u.ValidateRequest)); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if err := json.Unmarshal([]byte(result), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func (u *updateManager) getSecurityScanFeature(bnazaiCLI cli.Cli, orgID int32, clusterID int32) (map[string]interface{}, error) {

	clusterFeatureDetails, _, err := bnazaiCLI.Client().ClusterFeaturesApi.ClusterFeatureDetails(context.Background(), orgID, clusterID, featureName)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to retrieve the feature to update")
	}

	return clusterFeatureDetails.Spec, nil
}

func (u *updateManager) readUpdateReqFromFileOrStdin(filePath string, req *pipeline.UpdateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (u *updateManager) buildUpdateReqInteractively(req *pipeline.UpdateClusterFeatureRequest) error {
	var edit bool
	if err := survey.AskOne(&survey.Confirm{Message: "Edit the cluster feature update request in your text editor?"},
		&edit); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if !edit {
		return u.buildCustomAnchoreFeatureRequest(req)
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}

	var result string
	if err := survey.AskOne(
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true},
		&result,
		survey.WithValidator(u.ValidateRequest)); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	if err := json.Unmarshal([]byte(result), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func (u *updateManager) buildCustomAnchoreFeatureRequest(updateRequest *pipeline.UpdateClusterFeatureRequest) error {

	// get the type from the req
	securityFeatureSpec := new(SecurityScanFeatureSpec)
	if err := mapstructure.Decode(updateRequest.Spec, securityFeatureSpec); err != nil {
		return errors.WrapIf(err, "failed to decode the feature to update")
	}

	anchoreConfig, err := u.askForAnchoreConfig(&securityFeatureSpec.CustomAnchore)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore configuration details")
	}

	policy, err := u.askForPolicy(&securityFeatureSpec.Policy)
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	// todo whitelist updates not supported for now
	webhookConfig, err := u.askForWebHookConfig(&securityFeatureSpec.WebhookConfig)
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	securityScanFeatureRequest := new(SecurityScanFeatureSpec)
	securityScanFeatureRequest.CustomAnchore = *anchoreConfig
	securityScanFeatureRequest.Policy = *policy
	securityScanFeatureRequest.ReleaseWhiteList = securityFeatureSpec.ReleaseWhiteList
	securityScanFeatureRequest.WebhookConfig = *webhookConfig

	ssfMap, err := u.securityScanSpecAsMap(securityScanFeatureRequest)
	if err != nil {
		return errors.WrapIf(err, "failed to transform request to map")
	}

	updateRequest.Spec = ssfMap

	return nil
}
