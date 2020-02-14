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

package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type getOptions struct {
	clustercontext.Context
}

type getManager interface {
	ReadableName() string
	ServiceName() string
	WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{}
}

func newGetCommand(banzaiCLI cli.Cli, use string, mngr getManager) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"details", "show", "query"},
		Short:   fmt.Sprintf("Get details of the %s service for a cluster", mngr.ReadableName()),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(banzaiCLI, mngr, options, args, use)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("get %s cluster service details of", mngr.ReadableName()))

	return cmd
}

func runGet(
	banzaiCLI cli.Cli,
	m getManager,
	options getOptions,
	args []string,
	use string,
) error {
	if err := isServiceEnabled(context.Background(), banzaiCLI, use); err != nil {
		return errors.WrapIf(err, "failed to check service")
	}

	if err := options.Init(args...); err != nil {
		return errors.WrapIf(err, "failed to initialize options")
	}

	pipelineClient := banzaiCLI.Client()
	orgId := banzaiCLI.Context().OrganizationID()
	clusterId := options.ClusterID()

	details, resp, err := pipelineClient.IntegratedServicesApi.IntegratedServiceDetails(context.Background(), orgId, clusterId, m.ServiceName())

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("cluster service %q not found", m.ServiceName())
		return nil
	}

	if err != nil {
		cli.LogAPIError(fmt.Sprintf("get %s cluster service details", m.ReadableName()), err, resp.Request)
		log.Fatalf("could not get %s cluster service details: %v", m.ReadableName(), err)
		return err
	}

	// TODO (colin): refactor output writer, to use key/value pairs in each line
	if banzaiCLI.OutputFormat() == output.OutputFormatDefault {
		for name, tableData := range m.WriteDetailsTable(details) {
			var data interface{}
			var fields []string
			if banzaiCLI.OutputFormat() == output.OutputFormatDefault {
				data = tableData
				fields = make([]string, 0, len(tableData))
				for k := range tableData {
					fields = append(fields, k)
				}
				sort.Strings(fields)
			} else {
				data = details
			}

			ctx := &output.Context{
				Out:    banzaiCLI.Out(),
				Color:  banzaiCLI.Color(),
				Format: banzaiCLI.OutputFormat(),
				Fields: fields,
			}
			if _, err := fmt.Fprintln(ctx.Out, fmt.Sprintf(`
%s
----`, name)); err != nil {
				return errors.WrapIf(err, "failed to write table header")
			}

			if err := output.Output(ctx, data); err != nil {
				return errors.WrapIf(err, "failed to write output")
			}
		}
	} else {
		ctx := &output.Context{
			Out:    banzaiCLI.Out(),
			Color:  banzaiCLI.Color(),
			Format: banzaiCLI.OutputFormat(),
		}

		return output.Output(ctx, details)
	}

	return nil
}
