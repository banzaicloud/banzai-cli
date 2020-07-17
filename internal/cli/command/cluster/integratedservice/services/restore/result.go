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

package restore

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type resultOptions struct {
	clustercontext.Context

	restoreID int32
}

func newResultCommand(banzaiCli cli.Cli) *cobra.Command {
	options := resultOptions{}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g", "result", "details"},
		Short:   "Get restore result", // TODO (colin): add desc
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return showResult(banzaiCli, options)
		},
	}
	flags := cmd.Flags()
	flags.Int32VarP(&options.restoreID, "restoreId", "", 0, "Restore ID")
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "result")

	return cmd
}

func showResult(banzaiCli cli.Cli, options resultOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	if options.restoreID == 0 {
		if banzaiCli.Interactive() {
			restore, err := askRestore(client, orgID, clusterID)
			if err != nil {
				return errors.WrapIf(err, "failed to ask restore")
			}

			options.restoreID = restore.Id
		} else {
			return errors.NewWithDetails("invalid restore ID", "restoreID", options.restoreID)
		}
	}

	response, _, err := client.ArkRestoresApi.GetARKRestoreResuts(context.Background(), orgID, clusterID, options.restoreID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to get restore results", "clusterID", clusterID, "restoreID", options.restoreID)
	}

	ctx := &output.Context{
		Out:    banzaiCli.Out(),
		Color:  banzaiCli.Color(),
		Format: banzaiCli.OutputFormat(),
	}

	// log ARK errors
	ctx.Fields = []string{"Ark"}
	table := generateArkTable(response.Errors.Ark, "ERROR")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	// log cluster errors
	ctx.Fields = []string{"Cluster"}
	table = generateClusterTable(response.Errors.Cluster, "ERROR")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	// log namespace errors
	ctx.Fields = []string{"Namespace"}
	table = generateNamespacesTable(response.Errors.Namespaces, "ERROR")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	// log ARK warnings
	ctx.Fields = []string{"Ark"}
	table = generateArkTable(response.Warnings.Ark, "WARNING")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	// log cluster errors
	ctx.Fields = []string{"Cluster"}
	table = generateClusterTable(response.Warnings.Cluster, "WARNING")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	// log namespace warnings
	ctx.Fields = []string{"Namespace"}
	table = generateNamespacesTable(response.Warnings.Namespaces, "WARNING")
	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	return nil
}

func generateNamespacesTable(items map[string][]string, _type string) interface{} {
	var table interface{}
	if len(items) != 0 {
		type row struct {
			Namespace string
		}

		t := make([]row, 0, len(items))
		for namespace, item := range items {
			for _, e := range item {
				t = append(t, row{
					Namespace: fmt.Sprintf("%s (%s): %s", _type, namespace, e),
				})
			}
		}

		table = t
	}

	return table
}

func generateArkTable(items []string, _type string) interface{} {
	var table interface{}
	if len(items) != 0 {
		type row struct {
			Ark string
		}

		t := make([]row, 0, len(items))
		for _, e := range items {
			t = append(t, row{
				Ark: fmt.Sprintf("%s: %s", _type, e),
			})
		}

		table = t
	}

	return table
}

func generateClusterTable(items []string, _type string) interface{} {
	var table interface{}
	if len(items) != 0 {
		type row struct {
			Cluster string
		}

		t := make([]row, 0, len(items))
		for _, e := range items {
			t = append(t, row{
				Cluster: fmt.Sprintf("%s: %s", _type, e),
			})
		}

		table = t
	}

	return table
}
