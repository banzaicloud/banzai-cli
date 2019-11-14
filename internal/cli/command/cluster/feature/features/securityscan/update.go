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
)

type updateManager struct {
	baseManager
	specAssembler
}

func NewUpdateManager() features.UpdateManager {
	return updateManager{}
}

func (um updateManager) ValidateRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func (um updateManager) BuildRequestInteractively(banzaiCLI cli.Cli, updateClusterFeatureRequest *pipeline.UpdateClusterFeatureRequest, clusterCtx clustercontext.Context) error {

	// todo infer the cli directly to the manager instead
	um.specAssembler = specAssembler{banzaiCLI}

	if err := um.isFeatureEnabled(context.Background()); err != nil {
		return errors.WrapIf(err, "securityscan is not enabled")
	}

	featureSpec := SecurityScanFeatureSpec{}
	if err := mapstructure.Decode(updateClusterFeatureRequest.Spec, &featureSpec); err != nil {
		return errors.WrapIf(err, "failed to decode feature specification for update")
	}

	featureSpec, err := um.assembleFeatureSpec(context.Background(), banzaiCLI.Context().OrganizationID(), clusterCtx.ClusterID(), featureSpec)
	if err != nil {
		return errors.WrapIf(err, "failed to assemble feature specification")
	}

	featureSpecMap, err := um.securityScanSpecAsMap(&featureSpec)
	if err != nil {
		return errors.WrapIf(err, "failed to transform feature specification")
	}

	updateClusterFeatureRequest.Spec = featureSpecMap

	return nil
}
