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

package form

import (
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type migrateOptions struct {
	sourceConfigFile string
	targetConfigFile string
}

// NewMigrateCommand creates a new cobra.Command for `banzai form migrate`.
func NewMigrateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := migrateOptions{}

	cmd := &cobra.Command{
		Use:   "migrate SOURCE_FORM_CONFIG TARGET_FORM_CONFIG",
		Short: "Migrate form values from source config to target config",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			options.sourceConfigFile = args[0]
			options.targetConfigFile = args[1]
			err := runMigrate(banzaiCli, options)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd
}

func runMigrate(_ cli.Cli, options migrateOptions) error {
	source, err := readConfig(options.sourceConfigFile)
	if err != nil {
		return err
	}

	target, err := readConfig(options.targetConfigFile)
	if err != nil {
		return err
	}

	values := map[string]interface{}{}
	for _, group := range source.Form {
		for _, field := range group.Fields {
			values[field.Key] = field.Value
		}
	}

	for _, group := range target.Form {
		for _, field := range group.Fields {
			field.Value = values[field.Key]
		}
	}

	return writeConfig(options.targetConfigFile, target)
}
