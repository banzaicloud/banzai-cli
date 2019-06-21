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
	"errors"
	"io/ioutil"
	"os"

	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
)

const valuesDefault = "values.yaml"

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
	if err := utils.Unmarshal(kubeconfigContent, &config); err != nil {
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
