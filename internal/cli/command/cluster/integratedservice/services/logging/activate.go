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
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

type ActivateManager struct {
	baseManager
}

func (ActivateManager) BuildRequestInteractively(
	banzaiCLI cli.Cli,
	_ clustercontext.Context,
	cap map[string]interface{},
) (*pipeline.ActivateClusterFeatureRequest, error) {
	// get logging, tls and monitoring
	logging, err := askLogging(loggingSpec{
		Metrics: true, // TODO (colin): add monitoring integratedservice dependecy in v2
		TLS:     true,
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting settings options")
	}

	// get Loki
	loki, err := askLokiComponent(banzaiCLI, lokiSpec{
		Enabled: false,
		Ingress: ingressSpec{
			Enabled: false,
			Path:    "/loki",
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Loki options")
	}

	// get Cluster output
	clusterOutput, err := askClusterOutput(banzaiCLI, clusterOutputSpec{
		Enabled: true,
		Provider: providerSpec{
			Name: providerAmazonS3Key,
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "error during getting Cluster Output options")
	}

	return &pipeline.ActivateClusterFeatureRequest{
		Spec: map[string]interface{}{
			"logging":       logging,
			"loki":          loki,
			"clusterOutput": clusterOutput,
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

func askSecret(banzaiCLI cli.Cli, secretType, DefaultValue string, withSkipOption bool) (string, error) {

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

	if len(secrets) == 0 {
		// TODO (colin): add option to create new secret
		return "", nil
	}

	const skip = "skip"

	var secretName string
	var defaultSecretName string
	var secretLen = len(secrets)
	var secretIds = make(map[string]string, secretLen)
	if withSkipOption {
		defaultSecretName = skip
		secretLen = secretLen + 1
	}
	secretOptions := make([]string, secretLen)
	if withSkipOption {
		secretOptions[0] = skip
	}
	for i, s := range secrets {
		var idx = i
		if withSkipOption {
			idx = idx + 1
		}
		secretOptions[idx] = s.Name
		secretIds[s.Name] = s.Id
		if s.Id == DefaultValue {
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
		if b.Name == defaultValue {
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
		case providerAlibabaOSSKey:
			defaultProviderName = providerAlibabaOSSName
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
					QuestionBase: input.QuestionBase{},
					DefaultValue: defaultProviderName,
					Output:       &selectedProviderName,
				},
				Options: []string{providerAmazonS3Name, providerAzureName, providerGoogleGCSName, providerAlibabaOSSName},
			},
		}); err != nil {
			return nil, errors.WrapIf(err, "error during getting cluster output provider")
		}

		var providerOptions *providerSpec
		var err error
		switch selectedProviderName {
		case providerAlibabaOSSName:
			providerOptions, err = askOssOptions(banzaiCLI, defaults.Provider)
			if err != nil {
				return nil, errors.WrapIf(err, "failed to get OSS options")
			}
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

	bucket, err := askBuckets(banzaiCLI, amazonType, secretID, defaults.Name)
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

func askOssOptions(banzaiCLI cli.Cli, defaults providerSpec) (*providerSpec, error) {
	secretID, err := askSecret(banzaiCLI, alibabaType, defaults.SecretID, false)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get Alibaba secret")
	}

	bucket, err := askBuckets(banzaiCLI, alibabaType, secretID, defaults.Name)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get OSS buckets")
	}

	return &providerSpec{
		Name: providerAlibabaOSSKey,
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
