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
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type activateManager struct {
	baseManager
	specAssembler
}

func NewActivateManager() services.ActivateManager {
	return &activateManager{}
}

func (am activateManager) BuildRequestInteractively(banzaiCLI cli.Cli, clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {

	// todo infer the cli directly to the manager instead
	am.specAssembler = specAssembler{banzaiCLI}

	if err := am.isServiceEnabled(context.Background()); err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "securityscan is not enabled")
	}

	serviceSpec, err := am.assembleServiceSpec(context.Background(), banzaiCLI.Context().OrganizationID(), clusterCtx.ClusterID(), ServiceSpec{})
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to assemble integratedservice specification")
	}

	serviceSpecMap, err := am.securityScanSpecAsMap(&serviceSpec)
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to transform integratedservice specification")
	}

	return pipeline.ActivateIntegratedServiceRequest{Spec: serviceSpecMap}, nil
}

func (activateManager) ValidateSpec(spec map[string]interface{}) error {
	return nil
}

func (am activateManager) readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateIntegratedServiceRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func (am activateManager) securityScanSpecAsMap(spec *ServiceSpec) (map[string]interface{}, error) {
	// fill the structure of the config - make filling up the values easier
	if spec == nil {
		spec = &ServiceSpec{
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
