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

package expiry

import (
	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type UpdateManager struct {
	baseManager
}

func (UpdateManager) ValidateSpec(spec map[string]interface{}) error {
	return validateSpec(spec)
}

func (UpdateManager) BuildRequestInteractively(_ cli.Cli, req *pipeline.UpdateIntegratedServiceRequest, _ clustercontext.Context) error {
	var spec serviceSpec
	if err := mapstructure.Decode(req.Spec, &spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	date, err := askForDate(spec.Date)
	if err != nil {
		return errors.WrapIf(err, "error during getting date")
	}

	spec.Date = date
	if err := mapstructure.Decode(spec, &req.Spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	return nil
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{}
}
