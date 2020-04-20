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

package cluster

import (
	"context"
	"fmt"
	"io/ioutil"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type configOptions struct {
	clustercontext.Context
}

func NewConfigCommand(banzaiCli cli.Cli) *cobra.Command {
	options := configOptions{}

	cmd := &cobra.Command{
		Use:     "config [--cluster=ID | [--cluster-name=]NAME]",
		Aliases: []string{"co"},
		Short:   "Get K8S config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownloadConfig(banzaiCli, options, args)
		},
	}
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "config")

	return cmd
}

func runDownloadConfig(banzaiCli cli.Cli, options configOptions, args []string) error {
	if err := options.Init(args...); err != nil {
		return err
	}

	orgId := banzaiCli.Context().OrganizationID()
	id := options.ClusterID()

	config, _, err := banzaiCli.Client().ClustersApi.GetClusterConfig(context.Background(), orgId, id)
	if err != nil {
		return errors.WrapIf(err, "could not get cluster config")
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s.yaml", options.ClusterName()), []byte(config.Data), 0644)
	if err != nil {
		return errors.WrapIf(err, "failed to write initial repository config")
	}

	log.Infof("K8S config saved: %s", options.ClusterName())

	return nil
}
