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
	serviceName = "monitoring"

	ingressTypeGrafana      = "Grafana"
	ingressTypePrometheus   = "Prometheus"
	ingressTypeAlertmanager = "Alertmanager"
	ingressTypePushgateway  = "Pushgateway"

	passwordSecretType   = "password"
	htPasswordSecretType = "htpasswd"
	slackSecretType      = "slack"
	pagerDutySecretType  = "pagerduty"

	pagerDutyIntegrationEventApiV2 = "eventsApiV2"
	pagerDutyIntegrationPrometheus = "prometheus"

	alertmanagerProviderSlack             = "slack"
	alertmanagerProviderPagerDuty         = "pagerDuty"
	alertmanagerNotificationNameSlack     = "Slack"
	alertmanagerNotificationNamePagerDuty = "PagerDuty"

	pdIntegrationTypePrometheus      = "prometheus"
	pdIntegrationTypePrometheusName  = "Prometheus"
	pdIntegrationTypeEventsApiV2     = "eventsApiV2"
	pdIntegrationTypeEventsApiV2Name = "Events API V2"
)

type baseManager struct{}

func (baseManager) GetName() string {
	return serviceName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

func validateSpec(specObj map[string]interface{}) error {

	var spec serviceSpec

	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	return spec.Validate()
}
