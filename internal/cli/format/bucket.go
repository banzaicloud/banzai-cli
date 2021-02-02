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

	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

func BucketWrite(context formatContext, data interface{}) {
	bucketsWrite(context.Out(), context.OutputFormat(), context.Color(), data, []string{"Name", "Cloud", "Location", "Status"})
}

func DetailedBucketWrite(context formatContext, data interface{}, cloud string) {
	switch cloud {
	case input.CloudProviderAzure:
		AzureBucketWrite(context.Out(), context.OutputFormat(), context.Color(), data)
		return
	default:
		bucketsWrite(context.Out(), context.OutputFormat(), context.Color(), data, []string{"Name", "Cloud", "Location", "Status", "StatusMessage"})
	}
}

func AzureBucketWrite(out io.Writer, format string, color bool, data interface{}) {
	bucketsWrite(out, format, color, data, []string{"Name", "Cloud", "Location", "ResourceGroup", "StorageAccount", "Status", "StatusMessage"})
}

func bucketsWrite(out io.Writer, format string, color bool, data interface{}, fields []string) {
	ctx := &output.Context{
		Out:    out,
		Color:  color,
		Format: format,
		Fields: fields,
	}

	err := output.Output(ctx, data)
	if err != nil {
		log.Fatal(err)
	}
}
