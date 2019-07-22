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

package dns

import (
	"context"
	"strings"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewGetCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"details", "show", "query"},
		Short:   "Get details of the DNS feature for a cluster",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "get DNS cluster feature details of")

	return cmd
}

type getOptions struct {
	clustercontext.Context
}

const featureName = "dns"

func runGet(banzaiCli cli.Cli, options getOptions, args []string) error {
	if err := options.Init(args...); err != nil {
		return emperror.Wrap(err, "failed to initialize options")
	}

	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()

	details, resp, err := pipeline.ClusterFeaturesApi.ClusterFeatureDetails(context.Background(), orgId, clusterId, featureName)
	if err != nil {
		cli.LogAPIError("get DNS cluster feature details", err, resp.Request)
		log.Fatalf("could not get DNS cluster feature details: %v", err)
		return err
	}

	writeDNSFeatureDetails(banzaiCli, details)
	return nil
}

func writeDNSFeatureDetails(banzaiCli cli.Cli, details pipeline.ClusterFeatureDetails) error {
	var (
		data   interface{}
		fields []string
	)

	if banzaiCli.OutputFormat() == output.OutputFormatDefault {
		tableData := getTableData(details)

		data = tableData
		fields = make([]string, 0, len(tableData))
		for k := range tableData {
			fields = append(fields, k)
		}
	} else {
		data = details
	}

	ctx := &output.Context{
		Out:    banzaiCli.Out(),
		Color:  banzaiCli.Color(),
		Format: banzaiCli.OutputFormat(),
		Fields: fields,
	}

	return output.Output(ctx, data)
}

func getTableData(details pipeline.ClusterFeatureDetails) map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	if autodns, ok := getObj(details.Output, "autodns"); ok {
		if zone, ok := getStr(autodns, "zone"); ok {
			tableData["AutoDNS zone"] = zone
		}
		if clusterDomain, ok := getStr(autodns, "clusterDomain"); ok {
			tableData["AutoDNS cluster domain"] = clusterDomain
		}
	}

	if customDNS, ok := getObj(details.Spec, "customdns"); ok {
		if clusterDomain, ok := getObj(customDNS, "clusterDomain"); ok {
			tableData["CustomDNS cluster domain"] = clusterDomain
		}
		if domainFilters, ok := getList(customDNS, "domainFilters"); ok {
			filters := make([]string, 0, len(domainFilters))
			for _, f := range domainFilters {
				if s, ok := f.(string); ok {
					filters = append(filters, s)
				}
			}
			tableData["CustomDNS domain filters"] = strings.Join(filters, ",")
		}
		if provider, ok := getObj(customDNS, "provider"); ok {
			if name, ok := getStr(provider, "name"); ok {
				tableData["CustomDNS provider"] = name
			}
		}
	}

	return tableData
}

func getList(target map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := target[key]; ok {
		if list, ok := value.([]interface{}); ok {
			return list, true
		}
	}
	return nil, false
}

func getObj(target map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := target[key]; ok {
		if obj, ok := value.(map[string]interface{}); ok {
			return obj, true
		}
	}
	return nil, false
}

func getStr(target map[string]interface{}, key string) (string, bool) {
	if value, ok := target[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}
