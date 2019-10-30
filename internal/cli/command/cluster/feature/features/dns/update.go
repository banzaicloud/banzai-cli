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

package dns

import (
	"encoding/json"

	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type UpdateManager struct {
	baseManager
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{}
}

func (UpdateManager) BuildRequestInteractively(banzaiCLI cli.Cli, req *pipeline.UpdateClusterFeatureRequest) error {

	var spec specResponse
	if err := mapstructure.Decode(req.Spec, &spec); err != nil {
		return errors.WrapIf(err, "feature specification does not conform to schema")
	}

	externalDNS, err := buildExternalDNSFeatureRequest(banzaiCLI, defaults{
	})
	if err != nil {
		return errors.Wrap(err, "failed to build custom DNS feature request")
	}
	req.Spec = externalDNS

	return nil
}

func (UpdateManager) ValidateRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}
