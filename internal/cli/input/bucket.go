// Copyright © 2018 Banzai Cloud
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
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	AlibabaBucketNameRegex = "^[a-z]([-a-z0-9]{1,62}[a-z0-9])$"
	AmazonBucketNameRegex  = "^[a-z]([-a-z0-9]{1,62}[a-z0-9])$"
	AzureBucketNameRegex   = "^[a-z]([-a-z0-9]{1,61}[a-z0-9])$"
	GoogleBucketNameRegex  = "^[a-z]([-a-z0-9]{1,61}[a-z0-9])$"
	OracleBucketNameRegex  = "^[A-Za-z0-9-_.]{1,255}$"
	BucketNameRegex        = "^[\\w-.]{3,63}$"
)

func ValidateBucketName(cloud, name string) error {
	var regex string

	switch cloud {
	case CloudProviderAlibaba:
		regex = AlibabaBucketNameRegex
	case CloudProviderAmazon:
		regex = AmazonBucketNameRegex
	case CloudProviderAzure:
		regex = AzureBucketNameRegex
	case CloudProviderGoogle:
		regex = GoogleBucketNameRegex
	case CloudProviderOracle:
		regex = OracleBucketNameRegex
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	if !r.MatchString(name) {
		return errors.Errorf("invalid bucket name: '%s': must match regex: '%s'", name, r.String())
	}

	return nil
}

// BucketNameValidator validates bucket name for the specified cloud provider
func BucketNameValidator(cloud string) survey.Validator {
	return func(val interface{}) error {
		if str, ok := val.(string); ok {
			err := ValidateBucketName(cloud, str)
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf("cannot enforce length on response of type %v", reflect.TypeOf(val).Name())
		}
		return nil
	}
}
