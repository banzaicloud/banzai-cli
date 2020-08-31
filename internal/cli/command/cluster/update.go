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
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type updateOptions struct {
	file     string
	wait     bool
	interval int
	clustercontext.Context
}

// NewUpdateCommand creates a new cobra.Command for `banzai cluster update`.
func NewUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateOptions{}

	cmd := &cobra.Command{
		Use:          "update",
		Aliases:      []string{"u"},
		Short:        "Update a cluster",
		Long:         "Update cluster based on json stdin or interactive session",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Cluster update descriptor file")
	flags.BoolVarP(&options.wait, "wait", "w", false, "Wait for cluster update")
	flags.IntVarP(&options.interval, "interval", "i", 10, "Interval in seconds for polling cluster status")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "update")

	return cmd
}

func runUpdate(banzaiCli cli.Cli, options updateOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	err := options.Init()
	if err != nil {
		return err
	}

	id := options.ClusterID()

	var request pipeline.UpdateClusterRequest

	if banzaiCli.Interactive() {
		if cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgID, id); err != nil {
			return errors.WrapIf(err, "failed to get cluster details")
		} else {
			format.ClusterWrite(banzaiCli, cluster)
		}

		confirmed := false
		err := survey.AskOne(&survey.Confirm{Message: "Do you want to UPDATE the cluster?"}, &confirmed)
		if err != nil {
			return errors.WrapIf(err, "failed to read cluster update confirmation")
		}
		if !confirmed {
			return errors.New("update cancelled")
		}

		err = survey.AskOne(&survey.Input{Message: "Which version you wish to update to?"}, &request.Version)
		if err != nil {
			return errors.WrapIf(err, "failed to read cluster version")
		}
	} else {
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
		}

		log.Debugf("%d bytes read", len(raw))

		if err := utils.Unmarshal(raw, &request); err != nil {
			return errors.WrapIf(err, "failed to unmarshal update cluster request")
		}
	}

	log.Debugf("update request: %#v", request)

	_, err = client.ClustersApi.UpdateCluster(context.Background(), orgID, id, request)
	if err != nil {
		cli.LogAPIError("update cluster", err, request)
		return errors.WrapIf(err, "failed to update cluster")
	}

	log.Info("cluster is being updated")
	if options.wait {
		for {
			cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgID, id)
			if err != nil {
				cli.LogAPIError("get cluster", err, request)
			} else {
				format.ClusterShortWrite(banzaiCli, cluster)
				if cluster.Status != "UPDATING" {
					return nil
				}

				time.Sleep(time.Duration(options.interval) * time.Second)
			}
		}
	} else {
		log.Infof("you can check its status with the command `banzai cluster get %q`", options.ClusterName())
	}
	return nil
}
