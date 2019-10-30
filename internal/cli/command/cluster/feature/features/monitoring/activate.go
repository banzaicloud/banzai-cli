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
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"emperror.dev/errors"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/mitchellh/mapstructure"
)

type ActivateManager struct {
	baseManager
}

func (ActivateManager) BuildRequestInteractively(banzaiCLI cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {

	grafana, err := askGrafana(banzaiCLI, grafanaSpec{
		Enabled:    true,
		Dashboards: true,
		Ingress: baseIngressSpec{
			Enabled: true,
			Path:    "/grafana",
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Grafana options")
	}

	prometheus, err := askPrometheus(banzaiCLI, prometheusSpec{
		Enabled: true,
		Storage: storageSpec{
			Size:      100,
			Retention: "10d",
		},
		Ingress: ingressSpecWithSecret{
			baseIngressSpec: baseIngressSpec{
				Enabled: true,
				Path:    "/prometheus",
			},
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Prometheus options")
	}

	alertmanager, err := askAlertmanager(banzaiCLI, alertmanagerSpec{
		Enabled: true,
		Ingress: ingressSpecWithSecret{
			baseIngressSpec: baseIngressSpec{
				Enabled: true,
				Path:    "/alertmanager",
			},
		},
		Provider: map[string]interface{}{
			alertmanagerProviderSlack: slackSpec{
				Enabled:      false,
				SendResolved: true,
			},
			alertmanagerProviderPagerDuty: pagerDutySpec{
				Enabled:      false,
				SendResolved: true,
			},
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Alertmanager options")
	}

	pushgateway, err := askPushgateway(banzaiCLI, pushgatewaySpec{
		Enabled: true,
		Ingress: ingressSpecWithSecret{
			baseIngressSpec: baseIngressSpec{
				Enabled: false,
				Path:    "/pushgateway",
			},
		},
	})

	return &pipeline.ActivateClusterFeatureRequest{
		Spec: map[string]interface{}{
			"grafana":      grafana,
			"prometheus":   prometheus,
			"alertmanager": alertmanager,
			"pushgateway":  pushgateway,
			"exporters": exportersSpec{
				Enabled:          true,
				NodeExporter:     true,
				KubeStateMetrics: true,
			},
		},
	}, nil
}

func (ActivateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func askIngress(compType string, defaults baseIngressSpec) (*baseIngressSpec, error) {
	var isIngressEnabled bool
	var domain string
	var path string

	if err := doQuestions([]questionMaker{
		questionConfirm{
			questionBase: questionBase{
				message: fmt.Sprintf("Do you want to enable %s Ingress?", compType),
			},
			defaultValue: defaults.Enabled,
			output:       &isIngressEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, fmt.Sprintf("error during getting %s ingress enabled", compType))
	}

	if isIngressEnabled {
		var questions = []questionMaker{
			questionInput{
				questionBase: questionBase{
					message: fmt.Sprintf("Please provide %s Ingress domain:", compType),
					help:    "Leave empty to use cluster's IP",
				},
				defaultValue: defaults.Domain,
				output:       &domain,
			},
			questionInput{
				questionBase: questionBase{
					message: fmt.Sprintf("Please provide %s Ingress path:", compType),
				},
				defaultValue: defaults.Path,
				output:       &path,
			},
		}
		if err := doQuestions(questions); err != nil {
			return nil, errors.WrapIf(err, "error during asking ingress fields")
		}

	}
	return &baseIngressSpec{
		Enabled: isIngressEnabled,
		Domain:  domain,
		Path:    path,
	}, nil
}

func askGrafana(banzaiCLI cli.Cli, defaults grafanaSpec) (*grafanaSpec, error) {
	var isEnabled bool
	if err := doQuestions([]questionMaker{
		questionConfirm{
			questionBase: questionBase{
				message: "Do you want to enable Grafana?",
			},
			defaultValue: defaults.Enabled,
			output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Grafana enabled")
	}

	var result = &grafanaSpec{
		Enabled: isEnabled,
	}
	if isEnabled {
		var err error
		// secret
		result.SecretId, err = askSecret(banzaiCLI, passwordSecretType, defaults.SecretId)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting Grafana secret")
		}

		// ingress
		ingressSpec, err := askIngress("Grafana", defaults.Ingress)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting Grafana ingress options")
		}
		result.Ingress = *ingressSpec

		// default dashboards
		if err := doQuestions([]questionMaker{
			questionConfirm{
				questionBase: questionBase{
					message: "Do you want to add default dashboards to Grafana?",
				},
				defaultValue: defaults.Dashboards,
				output:       &result.Dashboards,
			},
		}); err != nil {
			return nil, errors.WrapIf(err, "error during getting default dashboards")
		}
	}

	return result, nil
}

func askPrometheus(banzaiCLI cli.Cli, defaults prometheusSpec) (*prometheusSpec, error) {
	var result = &prometheusSpec{
		Enabled: true,
	}

	// storage class, storage size and retention
	var storageSize = fmt.Sprint(defaults.Storage.Size)
	if err := doQuestions([]questionMaker{
		questionInput{
			questionBase: questionBase{
				message: "Please provide storage class name for Prometheus:",
				help:    "Leave empty to use default storage class",
			},
			defaultValue: defaults.Storage.Class,
			output:       &result.Storage.Class,
		},
		questionInput{
			questionBase: questionBase{
				message: "Please provide storage size for Prometheus:",
			},
			defaultValue: storageSize,
			output:       &storageSize,
		},
		questionInput{
			questionBase: questionBase{
				message: "Please provide retention for Prometheus:",
			},
			defaultValue: defaults.Storage.Retention,
			output:       &result.Storage.Retention,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Prometheus options")
	}

	storageSizeInt, err := strconv.ParseUint(storageSize, 10, 64)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to parse storage size")
	}
	result.Storage.Size = uint(storageSizeInt)

	// ingress
	ingressSpec, err := askIngress("Prometheus", defaults.Ingress.baseIngressSpec)
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Prometheus ingress options")
	}
	result.Ingress.baseIngressSpec = *ingressSpec

	if ingressSpec.Enabled {
		result.Ingress.SecretId, err = askSecret(banzaiCLI, htPasswordSecretType, defaults.Ingress.SecretId)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting secret for Prometheus ingress")
		}
	}

	return result, nil
}

func askAlertmanager(banzaiCLI cli.Cli, defaults alertmanagerSpec) (*alertmanagerSpec, error) {
	var isEnabled bool
	if err := doQuestions([]questionMaker{
		questionConfirm{
			questionBase: questionBase{
				message: "Do you want to enable Alertmanager?",
			},
			defaultValue: defaults.Enabled,
			output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Alertmanager enabled")
	}

	var result = &alertmanagerSpec{
		Enabled: isEnabled,
	}

	if isEnabled {
		result.Provider = map[string]interface{}{
			alertmanagerProviderSlack: slackSpec{
				Enabled: false,
			},
			alertmanagerProviderPagerDuty: pagerDutySpec{
				Enabled: false,
			},
		}

		// ask provider options
		var notificationProvider string
		var defaultNotificationProviderValue = alertmanagerNotificationNameSlack
		if pdProp, ok := defaults.Provider[alertmanagerProviderPagerDuty]; ok {
			var pd pagerDutySpec
			if err := mapstructure.Decode(pdProp, &pd); err == nil {
				if pd.Enabled {
					defaultNotificationProviderValue = alertmanagerNotificationNamePagerDuty
				}
			}
		}
		if err := doQuestions([]questionMaker{
			questionSelect{
				questionInput: questionInput{
					questionBase: questionBase{
						message: "Select notification provider",
					},
					defaultValue: defaultNotificationProviderValue,
					output:       &notificationProvider,
				},
				options: []string{alertmanagerNotificationNameSlack, alertmanagerNotificationNamePagerDuty},
			},
		}); err != nil {
			return nil, errors.WrapIf(err, "error during getting notification provider")
		}

		var err error
		switch notificationProvider {
		case alertmanagerNotificationNameSlack:
			result.Provider[alertmanagerProviderSlack], err = askNotificationProviderSlack(banzaiCLI, defaults.Provider[alertmanagerProviderSlack])
			if err != nil {
				return nil, errors.WrapIf(err, "error during getting Slack provider options")
			}
		case alertmanagerNotificationNamePagerDuty:
			result.Provider[alertmanagerProviderPagerDuty], err = askNotificationProviderPagerDuty(banzaiCLI, defaults.Provider[alertmanagerProviderPagerDuty])
			if err != nil {
				return nil, errors.WrapIf(err, "error during getting PagerDuty provider options")
			}
		default:
			return nil, errors.NewWithDetails("not supported provider type", "provider", notificationProvider)
		}

		// ask ingress
		ingressSpec, err := askIngress("Alertmanager", defaults.Ingress.baseIngressSpec)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting Alertmanager ingress options")
		}
		result.Ingress.baseIngressSpec = *ingressSpec

		if ingressSpec.Enabled {
			result.Ingress.SecretId, err = askSecret(banzaiCLI, htPasswordSecretType, defaults.Ingress.SecretId)
			if err != nil {
				return nil, errors.WrapIf(err, "error during getting secret for Alertmanager ingress")
			}
		}
	}

	return result, nil
}

func askSecret(banzaiCLI cli.Cli, secretType, defaultValue string) (string, error) {

	orgID := banzaiCLI.Context().OrganizationID()
	secrets, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(
		context.Background(),
		orgID,
		&pipeline.GetSecretsOpts{
			Type_: optional.NewString(secretType),
		},
	)
	if err != nil {
		return "", errors.WrapIf(err, "failed to get Vault secret(s)")
	}

	if len(secrets) == 0 {
		// TODO (colin): add option to create new secret
		return "", nil
	}

	const skip = "skip"

	var secretName string
	var defaultSecretName = skip
	secretOptions := make([]string, len(secrets)+1)
	secretIds := make(map[string]string, len(secrets))
	secretOptions[0] = skip
	for i, s := range secrets {
		secretOptions[i+1] = s.Name
		secretIds[s.Name] = s.Id
		if s.Id == defaultValue {
			defaultSecretName = s.Name
		}
	}

	if err := doQuestions([]questionMaker{questionSelect{
		questionInput: questionInput{
			questionBase: questionBase{
				message: "Provider secret:",
			},
			defaultValue: defaultSecretName,
			output:       &secretName,
		},
		options: secretOptions,
	}}); err != nil {
		return "", errors.WrapIf(err, "error during getting secret")
	}

	if secretName == skip {
		return "", nil
	}

	return secretIds[secretName], nil
}

func askNotificationProviderSlack(banzaiCLI cli.Cli, defaultsInterface interface{}) (*slackSpec, error) {
	var defaults slackSpec
	if err := mapstructure.Decode(defaultsInterface, &defaults); err != nil {
		return nil, errors.WrapIf(err, "failed to bind Slack config")
	}

	var err error
	var result = &slackSpec{
		Enabled: true,
	}
	result.SecretId, err = askSecret(banzaiCLI, slackSecretType, defaults.SecretId)
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Slack secret")
	}

	if err := doQuestions([]questionMaker{
		questionInput{
			questionBase: questionBase{
				message: "Provide Slack channel name for the alerts:",
			},
			output: &result.Channel,
		},
		questionConfirm{
			questionBase: questionBase{
				message: "Send resolved notifications as well",
			},
			defaultValue: defaults.SendResolved,
			output:       &result.SendResolved,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Slack options")
	}

	return result, nil
}

func askNotificationProviderPagerDuty(banzaiCLI cli.Cli, defaultsInterface interface{}) (*pagerDutySpec, error) {
	var defaults pagerDutySpec
	if err := mapstructure.Decode(defaultsInterface, &defaults); err != nil {
		return nil, errors.WrapIf(err, "failed to bind PagerDuty config")
	}

	var result = &pagerDutySpec{
		Enabled: true,
	}

	// ask for pd URL
	if err := doQuestions([]questionMaker{
		questionInput{
			questionBase: questionBase{
				message: "Provide PagerDuty service endpoint:",
			},
			defaultValue: defaults.Url,
			output:       &result.Url,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting PagerDuty url")
	}

	// ask for pd integration type
	var integrationType string
	var defaultIntegrationValue = pdIntegrationTypePrometheusName
	if defaults.IntegrationType == pdIntegrationTypeEventsApiV2 {
		defaultIntegrationValue = pdIntegrationTypeEventsApiV2Name
	}

	if err := doQuestions([]questionMaker{
		questionSelect{
			questionInput: questionInput{
				questionBase: questionBase{
					message: "Select PagerDuty integration type:",
				},
				defaultValue: defaultIntegrationValue,
				output:       &integrationType,
			},
			options: []string{pdIntegrationTypePrometheusName, pdIntegrationTypeEventsApiV2Name},
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting PagerDuty integration type")
	}

	switch integrationType {
	case pdIntegrationTypePrometheusName:
		result.IntegrationType = pdIntegrationTypePrometheus
	case pdIntegrationTypeEventsApiV2Name:
		result.IntegrationType = pdIntegrationTypeEventsApiV2
	default:
		return nil, errors.NewWithDetails("invalid integration type", "type", integrationType)
	}

	// ask for pd secret
	var err error
	result.SecretId, err = askSecret(banzaiCLI, pagerDutySecretType, defaults.SecretId)
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting PagerDuty secret")
	}

	if err := doQuestions([]questionMaker{
		questionConfirm{
			questionBase: questionBase{
				message: "Send resolved notifications as well",
			},
			defaultValue: defaults.SendResolved,
			output:       &result.SendResolved,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting PagerDuty send resolved option")
	}

	return result, nil
}

func askPushgateway(banzaiCLI cli.Cli, defaults pushgatewaySpec) (*pushgatewaySpec, error) {
	var isEnabled bool
	if err := doQuestions([]questionMaker{
		questionConfirm{
			questionBase: questionBase{
				message: "Do you want to enable Pushgateway?",
			},
			defaultValue: defaults.Enabled,
			output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Pushgateway enabled")
	}

	var result = &pushgatewaySpec{
		Enabled: isEnabled,
	}

	if isEnabled {
		// ask ingress
		ingressSpec, err := askIngress("Pushgateway", defaults.Ingress.baseIngressSpec)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting Pushgateway ingress options")
		}
		result.Ingress.baseIngressSpec = *ingressSpec

		if ingressSpec.Enabled {
			result.Ingress.SecretId, err = askSecret(banzaiCLI, htPasswordSecretType, defaults.Ingress.SecretId)
			if err != nil {
				return nil, errors.WrapIf(err, "error during getting secret for Pushgateway ingress")
			}
		}
	}

	return result, nil
}
