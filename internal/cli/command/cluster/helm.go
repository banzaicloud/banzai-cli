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

package cluster

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type helmOptions struct {
	Context
}

func NewHelmCommand(banzaiCli cli.Cli) *cobra.Command {
	options := helmOptions{}

	cmd := &cobra.Command{
		Use:    "_helm",
		Hidden: true,
		Short:  "Wrapper to download and execute the Helm version matching the Tiller on the cluster of the current kubecontext, in a Helm home that is synchronized with Pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHelm(banzaiCli, options, args)
		},
	}
	options.Context = NewClusterContext(cmd, banzaiCli, "get")

	return cmd
}

func runHelm(banzaiCli cli.Cli, options helmOptions, args []string) error {
	c := exec.Command("kubectl", "get", "deployment", "-n", "kube-system", "-o", "jsonpath={.items[0].spec.template.spec.containers[0].image}", "-l", "app=helm")
	out, err := c.Output()
	if err != nil {
		return emperror.Wrap(err, "failed to determine version of Tiller on the cluster")
	}
	parts := strings.Split(string(out), ":")
	if len(parts) != 2 {
		return errors.Errorf("failed to parse Tiller image name %q", out)
	}
	version := parts[1]

	fmt.Printf("https://get.helm.sh/helm-%s-%s-amd64.tar.gz", version, runtime.GOOS)

	return nil
}
