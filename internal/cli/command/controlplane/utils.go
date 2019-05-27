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

package controlplane

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/goph/emperror"
)

func unmarshal(raw []byte, data interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err == nil {
		return nil
	}

	// if can't decode as json, try to convert it from yaml first
	// use this method to prevent unmarshalling directly with yaml, for example to map[interface{}]interface{}
	converted, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return emperror.Wrap(err, "unmarshal")
	}

	decoder = json.NewDecoder(bytes.NewReader(converted))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return emperror.Wrap(err, "unmarshal")
	}

	return nil
}
