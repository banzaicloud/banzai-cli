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
	"github.com/mitchellh/mapstructure"
)

const (
	featureName          = "logging"
	htpasswordSecretType = "htpasswd"
	amazonType           = "amazon"
	azureType            = "azure"
	googleType           = "google"
	alibabaType          = "alibaba"

	providerAmazonS3Key    = "s3"
	providerAmazonS3Name   = "Amazon S3"
	providerGoogleGCSKey   = "gcs"
	providerGoogleGCSName  = "Google Cloud Storage"
	providerAlibabaOSSKey  = "oss"
	providerAlibabaOSSName = "Alibaba Object Storage"
	providerAzureKey       = "azure"
	providerAzureName      = "Azure Blob Storage"
)

type baseManager struct{}

func (baseManager) GetName() string {
	return featureName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

func validateSpec(specObj map[string]interface{}) error {
	var spec featureSpec
	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "integratedservice specification does not conform to schema")
	}

	return spec.Validate()
}
