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

package features

import (
	"context"
	"fmt"
	"log"
	"sort"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
	"github.com/spf13/cobra"
)

type getOptions struct {
	clustercontext.Context
}

type GetManager interface {
	// todo rename the method to getName
	GetCommandName() string
	WriteDetailsTable(pipeline.ClusterFeatureDetails) map[string]interface{}
}

func GetCommandFactory(banzaiCLI cli.Cli, manager GetManager, name string) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"details", "show", "query"},
		Short:   fmt.Sprintf("Get details of the %s feature for a cluster", name),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(banzaiCLI, manager, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("get %s cluster feature details of", name))

	return cmd
}

func runGet(
	banzaiCLI cli.Cli,
	m GetManager,
	options getOptions,
	args []string,
) error {
	if err := options.Init(args...); err != nil {
		return errors.WrapIf(err, "failed to initialize options")
	}

	pipelineClient := banzaiCLI.Client()
	orgId := banzaiCLI.Context().OrganizationID()
	clusterId := options.ClusterID()

	details, resp, err := pipelineClient.ClusterFeaturesApi.ClusterFeatureDetails(context.Background(), orgId, clusterId, m.GetCommandName())
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("get %s cluster feature details", m.GetCommandName()), err, resp.Request)
		log.Fatalf("could not get %s cluster feature details: %v", m.GetCommandName(), err)
		return err
	}

	var data interface{}
	var fields []string
	if banzaiCLI.OutputFormat() == output.OutputFormatDefault {
		tableData := m.WriteDetailsTable(details)

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

	return output.Output(ctx, data)
}
