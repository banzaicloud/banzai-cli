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
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

type controlPlaneInstallerOptions struct {
	installerTag  string
	pullInstaller bool
}

func (o *controlPlaneInstallerOptions) pullDockerImage() error {

	args := []string{
		"pull",
		fmt.Sprintf("banzaicloud/cp-installer:%s", o.installerTag),
	}

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func bindInstallerFlags(flags *flag.FlagSet, options *controlPlaneInstallerOptions) {
	flags.StringVarP(&options.installerTag, "image-tag", "", "latest", "Tag of banzaicloud/cp-installer Docker image to use")
	flags.BoolVarP(&options.pullInstaller, "image-pull", "", true, "Pull cp-installer image even if it's present locally")
}
