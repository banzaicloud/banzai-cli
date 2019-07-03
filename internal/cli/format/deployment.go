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

package format

import (
	"io"

	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

func deploymentWrite(out io.Writer, format string, color bool, data interface{}, fields []string) error {
	ctx := &output.Context{
		Out:    out,
		Color:  color,
		Format: format,
		Fields: fields,
	}

	if format == "json" || format == "yaml" {
		ctx.Fields = append(ctx.Fields, "Values", "Notes")
	}

	err := output.Output(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func DeploymentsWrite(out io.Writer, format string, color bool, data interface{}) error {
	fields := []string{"Namespace", "ReleaseName", "Status", "Version", "UpdatedAt", "CreatedAt", "ChartName", "ChartVersion"}

	return deploymentWrite(out, format, color, data, fields)
}

func DeploymentWrite(out io.Writer, format string, color bool, data interface{}) error {
	return DeploymentsWrite(out, format, color, []interface{}{data})
}

func DeploymentDeleteResponseWrite(out io.Writer, format string, color bool, data interface{}) error {
	fields := []string{"Name", "Status", "Message"}

	return deploymentWrite(out, format, color, []interface{}{data}, fields)
}

func DeploymentCreateUpdateResponseWrite(out io.Writer, format string, color bool, data interface{}) error {
	fields := []string{"ReleaseName", "Notes"}

	return deploymentWrite(out, format, color, []interface{}{data}, fields)
}

func HelmReposWrite(out io.Writer, format string, color bool, data interface{}) error {
	fields := []string{"Name", "Url"}

	return deploymentWrite(out, format, color, data, fields)
}
