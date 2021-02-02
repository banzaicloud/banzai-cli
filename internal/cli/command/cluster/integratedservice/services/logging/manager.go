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
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

type Manager struct {
	banzaiCLI cli.Cli
}

func NewManager(banzaiCLI cli.Cli) Manager {
	return Manager{
		banzaiCLI: banzaiCLI,
	}
}

func (Manager) ReadableName() string {
	return "Logging"
}

func (Manager) ServiceName() string {
	return "logging"
}

func (m Manager) BuildActivateRequestInteractively(clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {
	// get logging, tls and monitoring
	logging, err := askLogging(loggingSpec{
		Metrics: true, // TODO (colin): add monitoring integratedservice dependecy in v2
		TLS:     true,
	})
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "error during getting settings options")
	}

	// get Loki
	loki, err := askLokiComponent(m.banzaiCLI, lokiSpec{
		Enabled: false,
		Ingress: ingressSpec{
			Enabled: false,
			Path:    "/loki",
		},
	})
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "error during getting Loki options")
	}

	// get Cluster output
	clusterOutput, err := askClusterOutput(m.banzaiCLI, clusterOutputSpec{
		Enabled: true,
		Provider: providerSpec{
			Name: providerAmazonS3Key,
		},
	})
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "error during getting Cluster Output options")
	}

	return pipeline.ActivateIntegratedServiceRequest{
		Spec: map[string]interface{}{
			"logging":       logging,
			"loki":          loki,
			"clusterOutput": clusterOutput,
		},
	}, nil
}

func (m Manager) BuildUpdateRequestInteractively(clusterCtx clustercontext.Context, request *pipeline.UpdateIntegratedServiceRequest) error {
	var spec spec
	if err := mapstructure.Decode(request.Spec, &spec); err != nil {
		return errors.WrapIf(err, "integratedservice specification does not conform to schema")
	}

	// get logging, tls and monitoring
	logging, err := askLogging(spec.Logging)
	if err != nil {
		return errors.WrapIf(err, "error during getting settings options")
	}

	// get Loki
	loki, err := askLokiComponent(m.banzaiCLI, spec.Loki)
	if err != nil {
		return errors.WrapIf(err, "error during getting Loki options")
	}

	// get Cluster output
	clusterOutput, err := askClusterOutput(m.banzaiCLI, spec.ClusterOutput)
	if err != nil {
		return errors.WrapIf(err, "error during getting Cluster Output options")
	}

	request.Spec["logging"] = logging
	request.Spec["loki"] = loki
	request.Spec["clusterOutput"] = clusterOutput

	return nil
}

func (Manager) ValidateSpec(specObj map[string]interface{}) error {
	var spec spec
	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "integratedservice specification does not conform to schema")
	}

	return spec.Validate()
}

type outputResponse struct {
	Logging struct {
		OperatorVersion  string `mapstructure:"operatorVersion"`
		FluentdVersion   string `mapstructure:"fluentdVersion"`
		FluentbitVersion string `mapstructure:"fluentbitVersion"`
	} `mapstructure:"logging"`
	Loki struct {
		Url        string `mapstructure:"url"`
		Version    string `mapstructure:"version"`
		ServiceURL string `mapstructure:"serviceUrl"`
		SecretID   string `mapstructure:"secretId"`
	} `mapstructure:"loki"`
}

type TableData map[string]interface{}

func (Manager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	var output outputResponse
	if err := mapstructure.Decode(details.Output, &output); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return map[string]map[string]interface{}{
			"Logging": tableData,
		}
	}

	var spec spec
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return map[string]map[string]interface{}{
			"Logging": tableData,
		}
	}

	if spec.Loki.Enabled && spec.Loki.Ingress.Enabled {
		var secretID string
		if spec.Loki.Ingress.Enabled {
			secretID = spec.Loki.Ingress.SecretID
			if secretID == "" {
				secretID = output.Loki.SecretID
			}
		}
		var lokiTable = TableData{
			"url":        output.Loki.Url,
			"version":    output.Loki.Version,
			"serviceUrl": output.Loki.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Loki.Ingress.Path,
			"domain":     spec.Loki.Ingress.Domain,
		}
		tableData["Loki"] = lokiTable
	}

	tableData["Logging"] = TableData{
		"metrics": spec.Logging.Metrics,
		"TLS":     spec.Logging.TLS,
	}

	return map[string]map[string]interface{}{
		"Logging": tableData,
	}
}

func askLogging(defaults loggingSpec) (*loggingSpec, error) {
	var isTlS bool
	var isMetrics bool
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionConfirm{
			QuestionBase: input.QuestionBase{
				Message: "Do you want to enable TLS?",
			},
			DefaultValue: defaults.TLS,
			Output:       &isTlS,
		},
		input.QuestionConfirm{
			QuestionBase: input.QuestionBase{
				Message: "Do you want to enable Metrics?",
			},
			DefaultValue: defaults.Metrics,
			Output:       &isMetrics,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting logging information")
	}

	return &loggingSpec{
		Metrics: isMetrics,
		TLS:     isTlS,
	}, nil
}

func askLokiComponent(banzaiCLI cli.Cli, defaults lokiSpec) (*lokiSpec, error) {
	var isEnabled bool
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionConfirm{
			QuestionBase: input.QuestionBase{
				Message: "Do you want to enable Loki?",
			},
			DefaultValue: defaults.Enabled,
			Output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting Loki enabled")
	}

	var result = &lokiSpec{
		Enabled: isEnabled,
	}
	if isEnabled {
		ingress, err := askIngress("Loki", defaults.Ingress)
		if err != nil {
			return nil, errors.WrapIf(err, "error during getting Loki Options")
		}

		if ingress.Enabled {
			ingress.SecretID, err = askSecret(banzaiCLI, htpasswordSecretType, defaults.Ingress.SecretID, true)
			if err != nil {
				return nil, errors.WrapIf(err, "error during getting Loki secret")
			}
		}

		result.Ingress = *ingress
	}

	return result, nil
}

func askIngress(componentType string, defaults ingressSpec) (*ingressSpec, error) {
	var isEnabled bool

	var domain string
	var path string

	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionConfirm{
			QuestionBase: input.QuestionBase{
				Message: fmt.Sprintf("Do you want to enable %s Ingress?", componentType),
			},
			DefaultValue: defaults.Enabled,
			Output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, fmt.Sprintf("error during getting %s ingress enabled", componentType))
	}

	if isEnabled {
		if err := input.DoQuestions([]input.QuestionMaker{
			input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: fmt.Sprintf("Please provide %s Ingress domain:", componentType),
					Help:    "Leave empty to use cluster's IP",
				},
				DefaultValue: defaults.Domain,
				Output:       &domain,
			},
			input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: fmt.Sprintf("Please provide %s Ingress path:", componentType),
				},
				DefaultValue: defaults.Path,
				Output:       &path,
			},
		}); err != nil {
			return nil, errors.WrapIf(err, "error during asking ingress fields")
		}
	}

	return &ingressSpec{
		Enabled: isEnabled,
		Domain:  domain,
		Path:    path,
	}, nil
}

func askSecret(banzaiCLI cli.Cli, secretType, defaultValue string, withSkipOption bool) (string, error) {
	orgID := banzaiCLI.Context().OrganizationID()
	secrets, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(
		context.Background(),
		orgID,
		&pipeline.GetSecretsOpts{
			Type_: optional.NewString(secretType),
		},
	)
	if err != nil {
		return "", errors.WrapIfWithDetails(err, "failed to get secret(s)", "secretType", secretType)
	}

	// hide hidden secrets
	var finalSecrets []pipeline.SecretItem
	for _, s := range secrets {
		if !isSecretHidden(s) {
			finalSecrets = append(finalSecrets, s)
		}
	}

	if len(finalSecrets) == 0 {
		// TODO (colin): add option to create new secret
		return "", nil
	}

	const skip = "skip"

	var secretName string
	var defaultSecretName string
	var secretLen = len(finalSecrets)
	var secretIds = make(map[string]string, secretLen)
	if withSkipOption {
		defaultSecretName = skip
		secretLen = secretLen + 1
	}
	secretOptions := make([]string, secretLen)
	if withSkipOption {
		secretOptions[0] = skip
	}
	for i, s := range finalSecrets {
		var idx = i
		if withSkipOption {
			idx = idx + 1
		}
		secretOptions[idx] = s.Name
		secretIds[s.Name] = s.Id
		if s.Id == defaultValue || (defaultValue == "" && i == 0 && !withSkipOption) {
			defaultSecretName = s.Name
		}
	}

	if err := input.DoQuestions([]input.QuestionMaker{input.QuestionSelect{
		QuestionInput: input.QuestionInput{
			QuestionBase: input.QuestionBase{
				Message: "Provider secret:",
			},
			DefaultValue: defaultSecretName,
			Output:       &secretName,
		},
		Options: secretOptions,
	}}); err != nil {
		return "", errors.WrapIf(err, "error during getting secret")
	}

	if secretName == skip {
		return "", nil
	}

	return secretIds[secretName], nil
}

func askBuckets(banzaiCLI cli.Cli, bucketType, secretID, defaultValue string) (*pipeline.BucketInfo, error) {
	orgID := banzaiCLI.Context().OrganizationID()
	buckets, _, err := banzaiCLI.Client().StorageApi.ListObjectStoreBuckets(
		context.Background(),
		orgID,
		&pipeline.ListObjectStoreBucketsOpts{
			SecretId:  optional.NewString(secretID),
			CloudType: optional.NewString(bucketType),
		},
	)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to get bucket(s)", "secretID", secretID, "cloud", bucketType)
	}

	if len(buckets) == 0 {
		return nil, errors.New("there's no buckets")
	}

	var defaultBucketName string
	secretOptions := make([]string, len(buckets))
	for i, b := range buckets {
		secretOptions[i] = b.Name
		if b.Name == defaultValue || (defaultValue == "" && i == 0) {
			defaultBucketName = b.Name
		}
	}

	var bucketName string
	if err := input.DoQuestions([]input.QuestionMaker{input.QuestionSelect{
		QuestionInput: input.QuestionInput{
			QuestionBase: input.QuestionBase{
				Message: "Bucket name:",
			},
			DefaultValue: defaultBucketName,
			Output:       &bucketName,
		},
		Options: secretOptions,
	}}); err != nil {
		return nil, errors.WrapIf(err, "error during getting secret")
	}

	var bucket pipeline.BucketInfo
	for _, b := range buckets {
		if b.Name == bucketName {
			bucket = b
			return &bucket, nil
		}
	}

	return nil, errors.New("missing bucket")
}

func askClusterOutput(banzaiCLI cli.Cli, defaults clusterOutputSpec) (*clusterOutputSpec, error) {
	var isEnabled bool
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionConfirm{
			QuestionBase: input.QuestionBase{
				Message: "Do you want to enable cluster output?",
			},
			DefaultValue: defaults.Enabled,
			Output:       &isEnabled,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "error during getting cluster output enabled")
	}

	var result = clusterOutputSpec{
		Enabled: isEnabled,
	}

	if isEnabled {
		var defaultProviderName string
		switch defaults.Provider.Name {
		case providerGoogleGCSKey:
			defaultProviderName = providerGoogleGCSName
		case providerAzureKey:
			defaultProviderName = providerAzureName
		default:
			defaultProviderName = providerAmazonS3Name
		}

		var selectedProviderName string
		if err := input.DoQuestions([]input.QuestionMaker{
			input.QuestionSelect{
				QuestionInput: input.QuestionInput{
					QuestionBase: input.QuestionBase{
						Message: "Select log storage provider:",
					},
					DefaultValue: defaultProviderName,
					Output:       &selectedProviderName,
				},
				Options: []string{providerAmazonS3Name, providerAzureName, providerGoogleGCSName},
			},
		}); err != nil {
			return nil, errors.WrapIf(err, "error during getting cluster output provider")
		}

		var providerOptions *providerSpec
		var err error
		switch selectedProviderName {
		case providerAzureName:
			providerOptions, err = askAzureOptions(banzaiCLI, defaults.Provider)
			if err != nil {
				return nil, errors.WrapIf(err, "failed to get Azure options")
			}
		case providerGoogleGCSName:
			providerOptions, err = askGCSOptions(banzaiCLI, defaults.Provider)
			if err != nil {
				return nil, errors.WrapIf(err, "failed to get GCS options")
			}
		default:
			providerOptions, err = askS3Options(banzaiCLI, defaults.Provider)
			if err != nil {
				return nil, errors.WrapIf(err, "failed to get S3 options")
			}
		}

		result.Provider = *providerOptions
	}

	return &result, nil
}

func askS3Options(banzaiCLI cli.Cli, defaults providerSpec) (*providerSpec, error) {
	secretID, err := askSecret(banzaiCLI, amazonType, defaults.SecretID, false)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get Amazon secret")
	}

	bucket, err := askBuckets(banzaiCLI, amazonType, secretID, defaults.Bucket.Name)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get S3 buckets")
	}

	return &providerSpec{
		Name: providerAmazonS3Key,
		Bucket: bucketSpec{
			Name: bucket.Name,
		},
		SecretID: secretID,
	}, nil
}

func askGCSOptions(banzaiCLI cli.Cli, defaults providerSpec) (*providerSpec, error) {
	secretID, err := askSecret(banzaiCLI, googleType, defaults.SecretID, false)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get Google secret")
	}

	bucket, err := askBuckets(banzaiCLI, googleType, secretID, defaults.Name)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get GCS buckets")
	}

	return &providerSpec{
		Name: providerGoogleGCSKey,
		Bucket: bucketSpec{
			Name: bucket.Name,
		},
		SecretID: secretID,
	}, nil
}

func askAzureOptions(banzaiCLI cli.Cli, defaults providerSpec) (*providerSpec, error) {
	secretID, err := askSecret(banzaiCLI, azureType, defaults.SecretID, false)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get Azure secret")
	}

	bucket, err := askBuckets(banzaiCLI, azureType, secretID, defaults.Name)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get Azure buckets")
	}

	return &providerSpec{
		Name: providerAzureKey,
		Bucket: bucketSpec{
			Name:           bucket.Name,
			ResourceGroup:  bucket.Aks.ResourceGroup,
			StorageAccount: bucket.Aks.StorageAccount,
		},
		SecretID: secretID,
	}, nil
}

func isSecretHidden(s pipeline.SecretItem) bool {
	for _, t := range s.Tags {
		if t == "banzai:hidden" {
			return true
		}
	}
	return false
}
