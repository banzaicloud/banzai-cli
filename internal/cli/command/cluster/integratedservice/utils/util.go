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

package utils

import (
	"net/http"
	"sort"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2/core"
)

// helper type alias for id -> name maps
type IdToNameMap = map[string]string

func Names(sm IdToNameMap) []string {
	names := make([]string, len(sm))
	i := 0
	for _, name := range sm {
		names[i] = name
		i = i + 1
	}
	sort.Strings(names)
	return names
}

func NameForID(sm IdToNameMap, idOf string) string {
	for id, n := range sm {
		if id == idOf {
			return n
		}
	}
	return ""
}

// returns the first value in the map
func GetFirstName(sm IdToNameMap) string {
	for _, name := range sm {
		return name
	}
	return ""
}

// NameToIDTransformer returns a survey.Transformer that transform the map value into it's key
// todo remove the direct dependency on the survey api
func NameToIDTransformer(sm IdToNameMap) func(name interface{}) interface{} {
	return func(name interface{}) interface{} {
		for id, n := range sm {
			if n == name.(core.OptionAnswer).Value {
				return core.OptionAnswer{
					Value: id,
				}
			}
		}
		return nil
	}
}

// checks for errors and unexpected error code in the response data
func CheckCallResults(r *http.Response, err error) error {
	if err != nil {
		return errors.WrapIf(err, "failure during callout to pipeline")
	}
	if r.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code: %d, status: %s", r.StatusCode, r.Status)
	}
	return nil
}
