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
	"reflect"
	"strconv"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

// InputNumberValidator validates an integer number input, between min and max values
func InputNumberValidator(min, max int) survey.Validator {
	return func(val interface{}) error {
		if str, ok := val.(string); ok {
			i, err := strconv.Atoi(str)
			if err != nil {
				return errors.Wrap(err, "invalid input number")
			}
			if i > max {
				return errors.Errorf("value should be < %v", max)
			}
			if i <= min {
				return errors.Errorf("value should be > %v", min)
			}
		} else {
			return errors.Errorf("cannot enforce string on response of type %v", reflect.TypeOf(val).Name())
		}
		return nil
	}
}
