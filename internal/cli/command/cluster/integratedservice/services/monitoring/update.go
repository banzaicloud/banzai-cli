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

package monitoring

import (
	"encoding/json"

	"emperror.dev/errors"

	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type UpdateManager struct {
	baseManager
}

func (UpdateManager) ValidateRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}

func (UpdateManager) BuildRequestInteractively(banzaiCLI cli.Cli, req *pipeline.UpdateClusterFeatureRequest, clusterCtx clustercontext.Context) error {

	var spec serviceSpec
	if err := mapstructure.Decode(req.Spec, &spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	grafana, err := askGrafana(banzaiCLI, spec.Grafana)
	if err != nil {
		return errors.WrapIf(err, "error during getting Grafana options")
	}

	prometheus, err := askPrometheus(banzaiCLI, spec.Prometheus)
	if err != nil {
		return errors.WrapIf(err, "error during getting Prometheus options")
	}

	alertmanager, err := askAlertmanager(banzaiCLI, spec.Alertmanager)
	if err != nil {
		return errors.WrapIf(err, "error during getting Alertmanager options")
	}

	pushgateway, err := askPushgateway(banzaiCLI, spec.Pushgateway)
	if err != nil {
		return errors.WrapIf(err, "error during getting Pushgateway options")
	}

	req.Spec["grafana"] = grafana
	req.Spec["prometheus"] = prometheus
	req.Spec["alertmanager"] = alertmanager
	req.Spec["pushgateway"] = pushgateway

	return nil
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{}
}
