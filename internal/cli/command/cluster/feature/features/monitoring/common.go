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
	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"
)

const (
	featureName = "monitoring"

	ingressTypeGrafana      = "Grafana"
	ingressTypePrometheus   = "Prometheus"
	ingressTypeAlertmanager = "Alertmanager"

	passwordSecretType = "password"
)

func NewDeactivateManager() *ActivateManager {
	return &ActivateManager{}
}

func validateSpec(specObj map[string]interface{}) error {

	var spec featureSpec

	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "feature specification does not conform to schema")
	}

	// Grafana spec validation
	if spec.Grafana.Enabled && spec.Grafana.Public.Enabled {
		if err := spec.Grafana.Public.Validate(ingressTypeGrafana); err != nil {
			return err
		}
	}

	// Prometheus spec validation
	if spec.Prometheus.Enabled && spec.Prometheus.Public.Enabled {
		if err := spec.Prometheus.Public.Validate(ingressTypePrometheus); err != nil {
			return err
		}
	}

	// Alertmanager spec validation
	if spec.Alertmanager.Enabled && spec.Alertmanager.Public.Enabled {
		if err := spec.Alertmanager.Public.Validate(ingressTypeAlertmanager); err != nil {
			return err
		}
	}

	return nil
}
