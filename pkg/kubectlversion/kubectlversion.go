// Copyright © 2020 Banzai Cloud
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

package kubectlversion

import (
	"encoding/json"
	"os/exec"

	"emperror.dev/errors"
	"github.com/Masterminds/semver"
)

type clientVersion struct {
	GitVersion string `json:"gitVersion,omitempty"`
}

type kubectlVersionOutput struct {
	ClientVersion clientVersion `json:"clientVersion,omitempty"`
}

func getLocalKubectlVersion() (clientVersion, error) {
	c := exec.Command("kubectl", "version", "--client=true", "-o", "json")
	out, err := c.Output()
	if err != nil {
		return clientVersion{}, errors.WrapIf(err, "failed to determine kubectl version")
	}

	var parsed kubectlVersionOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return clientVersion{}, errors.WrapIf(err, "failed to parse kubectl version")
	}

	return parsed.ClientVersion, nil
}

func LessThan(v string) (bool, error) {
	inputVersion, err := semver.NewVersion(v)
	if err != nil {
		return false, errors.WrapIf(err, "failed to parse kubectl input version")
	}

	localKubectlVersion, err := getLocalKubectlVersion()
	if err != nil {
		return false, errors.WrapIf(err, "failed to get local kubectl version")
	}

	localVersion, err := semver.NewVersion(localKubectlVersion.GitVersion)
	if err != nil {
		return false, errors.WrapIf(err, "failed to parse kubectl local semantic version")
	}

	return localVersion.LessThan(inputVersion), nil
}
