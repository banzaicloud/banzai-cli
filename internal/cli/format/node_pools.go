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

package format

import (
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
	log "github.com/sirupsen/logrus"
)

// NodePoolsWrite writes a node pool list to the output.
func NodePoolsWrite(context formatContext, data interface{}) {
	ctx := &output.Context{
		Out:    context.Out(),
		Color:  context.Color(),
		Format: context.OutputFormat(),
		Fields: []string{"Name", "Size", "Autoscaling", "MinimumSize", "MaximumSize", "VolumeEncryption", "VolumeSize", "InstanceType", "Image", "SpotPrice", "SubnetID", "SecurityGroups", "Status", "StatusMessage"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		log.Fatal(err)
	}
}
