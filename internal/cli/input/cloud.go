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

package input

import (
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

const (
	CloudProviderAmazon = "amazon"
	CloudProviderAzure  = "azure"
	CloudProviderGoogle = "google"
)

// AskCloud asks for cloud provider
func AskCloud() (string, error) {
	var cloud string

	err := survey.AskOne(&survey.Select{Message: "Cloud:", Options: []string{
		CloudProviderAmazon,
		CloudProviderAzure,
		CloudProviderGoogle,
	}}, &cloud, survey.WithValidator(survey.Required))
	if err != nil {
		return cloud, errors.WrapIf(err, "failed to select cloud")
	}

	return cloud, nil
}

// IsCloudProviderSupported checks whether the given cloud provider is supported
func IsCloudProviderSupported(cloud string) error {
	switch cloud {
	case CloudProviderAmazon, CloudProviderAzure, CloudProviderGoogle:
	default:
		return errors.Errorf("invalid cloud provider specified: %s", cloud)
	}

	return nil
}
