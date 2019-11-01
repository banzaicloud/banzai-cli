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

package dns

import (
	"github.com/AlecAivazis/survey/v2/core"
)

// helper type alias for id -> name maps
type idToNameMap = map[string]string

func Names(sm idToNameMap) []string {
	names := make([]string, len(sm))
	i := 0
	for _, name := range sm {
		names[i] = name
		i = i + 1
	}
	return names
}

func NameForID(sm idToNameMap, idOf string) string {
	for id, n := range sm {
		if id == idOf {
			return n
		}
	}
	return ""
}

// nameToIDTransformer returns a survey.Transformer that transform the map value into it's key
// todo remove the direct dependency on the survey api
func nameToIDTransformer(sm idToNameMap) func(name interface{}) interface{} {
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
