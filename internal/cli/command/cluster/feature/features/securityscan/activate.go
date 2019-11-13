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
	return &activateManager{}
}

func (am activateManager) BuildRequestInteractively(banzaiCLI cli.Cli, clusterCtx clustercontext.Context) (*pipeline.ActivateClusterFeatureRequest, error) {
	var req pipeline.ActivateClusterFeatureRequest
	// todo infer the cli directly to the manager instead
	am.specAssembler = specAssembler{banzaiCLI}

	if err := am.isFeatureEnabled(context.Background()); err != nil {
		return nil, errors.WrapIf(err, "securityscan is not enabled")
	}

	if err := am.buildAnchoreConfigSpec(context.Background(), banzaiCLI.Context().OrganizationID(), clusterCtx.ClusterID(), &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (am activateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func (am activateManager) readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (am activateManager) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
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

func (am activateManager) buildAnchoreConfigSpec(ctx context.Context, orgID int32, clusterID int32, activateRequest *pipeline.ActivateClusterFeatureRequest) error {

	// todo uncomment this for supporting custom anchore
	//anchoreConfig, err := am.askForAnchoreConfig(nil)
	//if err != nil {
	//	return errors.WrapIf(err, "failed to read Anchore configuration details")
	//}

	policy, err := am.askForPolicy(ctx, orgID, clusterID, policySpec{})
	if err != nil {
		return errors.WrapIf(err, "failed to read Anchore Policy configuration details")
	}

	// todo uncomment this for supporting whitelist management
	//whiteLists, err := am.askForWhiteLists()
	//if err != nil {
	//	return errors.WrapIf(err, "failed to read whitelists")
	//}

	webhookConfig, err := am.askForWebHookConfig(ctx, orgID, clusterID, webHookConfigSpec{})
	if err != nil {
		return errors.WrapIf(err, "failed to read webhook configuration")
	}

	featureSpec := SecurityScanFeatureSpec{
		Policy:        policy,
		WebhookConfig: webhookConfig,
	}

	ssfMap, err := am.securityScanSpecAsMap(&featureSpec)
	if err != nil {
		return errors.WrapIf(err, "failed to transform request to map")
	}

	activateRequest.Spec = ssfMap

	return nil
}
