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

package logging

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/mitchellh/mapstructure"
)

type UpdateManager struct {
	baseManager
}

func (UpdateManager) ValidateSpec(spec map[string]interface{}) error {
	return validateSpec(spec)
}

func (UpdateManager) BuildRequestInteractively(banzaiCLI cli.Cli, req *pipeline.UpdateIntegratedServiceRequest, clusterCtx clustercontext.Context) error {
	var spec spec
	if err := mapstructure.Decode(req.Spec, &spec); err != nil {
		return errors.WrapIf(err, "integratedservice specification does not conform to schema")
	}

	// get logging, tls and monitoring
	logging, err := askLogging(spec.Logging)
	if err != nil {
		return errors.WrapIf(err, "error during getting settings options")
	}

	// get Loki
	loki, err := askLokiComponent(banzaiCLI, spec.Loki)
	if err != nil {
		return errors.WrapIf(err, "error during getting Loki options")
	}

	// get Cluster output
	clusterOutput, err := askClusterOutput(banzaiCLI, spec.ClusterOutput)
	if err != nil {
		return errors.WrapIf(err, "error during getting Cluster Output options")
	}

	req.Spec["logging"] = logging
	req.Spec["loki"] = loki
	req.Spec["clusterOutput"] = clusterOutput

	return nil
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{}
}
