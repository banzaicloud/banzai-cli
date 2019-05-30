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
	"errors"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
)

const valuesDefault = "values.yaml"

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
		return emperror.Wrap(err, "unmarshal YAML")
	}

	decoder = json.NewDecoder(bytes.NewReader(converted))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return emperror.Wrap(err, "unmarshal JSON")
	}

	return nil
}

// copyKubeconfig copies current Kubeconfig to the named file to a place where it is more likely that it can be mounted to DfM
func copyKubeconfig(kubeconfigName string) error {
	kubeconfigSource := os.Getenv("KUBECONFIG")
	if kubeconfigSource == "" {
		kubeconfigSource = os.Getenv("HOME") + "/.kube/config"
	}

	kubeconfigContent, err := ioutil.ReadFile(kubeconfigSource)
	if err != nil {
		return emperror.With(emperror.Wrapf(err, "failed to read kubeconfig %q", kubeconfigSource), "path", kubeconfigSource)
	}

	config := map[string]interface{}{}
	if err := unmarshal(kubeconfigContent, &config); err != nil {
		return emperror.Wrapf(err, "failed to parse kubeconfig %q", kubeconfigSource)
	}

	currentContext := config["current-context"]
	if currentContext == nil {
		return errors.New("can't find current context in kubeconfig")
	}

	log.Infof("Current Kubernetes context: %s", currentContext)

	if err := ioutil.WriteFile(kubeconfigName, kubeconfigContent, 0600); err != nil {
		return emperror.With(emperror.Wrapf(err, "failed to write temporary file %q", kubeconfigName), "path", kubeconfigName)
	}

	return nil
}
