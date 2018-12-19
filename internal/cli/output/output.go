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

package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/banzaicloud/banzai-cli/pkg/formatting"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Context contains parameters for formatting data.
type Context struct {
	Out    io.Writer
	Color  bool
	Format string
	Fields []string
}

// Output writes data in a specific format.
func Output(ctx *Context, data interface{}) error {
	switch ctx.Format {
	case "json":
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal output")
		}

		_, err = fmt.Fprintf(ctx.Out, "%s\n", bytes)

		return errors.Wrap(err, "cannot write output")

	case "yaml":
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return errors.Wrap(err, "cannot marshal output")
		}

		_, err = fmt.Fprintf(ctx.Out, "%s\n", bytes)

		return errors.Wrap(err, "cannot write output")

	default:
		table := formatting.NewTable(data, ctx.Fields)
		formatted := table.Format(ctx.Color)

		_, err := fmt.Fprintln(ctx.Out, formatted)

		return errors.Wrap(err, "cannot write output")
	}
}
