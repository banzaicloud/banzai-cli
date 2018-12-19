// Copyright © 2018 Banzai Cloud
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
	log "github.com/sirupsen/logrus"
)

// SecretWrite writes a secret list to the output.
func SecretWrite(out io.Writer, format string, color bool, data interface{}) {
	ctx := &output.Context{
		Out:    out,
		Color:  color,
		Format: format,
		Fields: []string{"Id", "Name", "Type", "UpdatedBy", "Tags"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		log.Fatal(err)
	}
}
