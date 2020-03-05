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

package controlplane

import (
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type migrateOptions struct {
	*cpContext
}

// NewMigrateCommand creates a new cobra.Command for `banzai pipeline up`.
func NewMigrateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := migrateOptions{}

	cmd := &cobra.Command{
		Use:     "migrate",
		Aliases: []string{"migrate"},
		Short:   "Migrate Banzai Cloud Pipeline Database",
		Long:    `Migrate database of Banzai Cloud Pipeline.` + initLongDescription,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runMigrate(&options, banzaiCli)
		},
	}

	options.cpContext = NewContext(cmd, banzaiCli)

	return cmd
}

func runMigrate(options *migrateOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	if banzaiCli.Interactive() {
		var migrate bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to MIGRATE the controlplane database now?",
				Default: true,
			},
			&migrate,
		)

		if !migrate {
			return errors.New("controlplane database migration cancelled")
		}
	}

	if err := applyMigrateModule(options.cpContext, map[string]string{}); err != nil {
		return err
	}
	return nil
}

func applyMigrateModule(options *cpContext, env map[string]string) error {
	log.Info("upgrading database...")
	targets := []string{"module.database.module.upgrade"}
	if err := runTerraform("apply", options, env, targets...); err != nil {
		return errors.WrapIf(err, "failed to upgrade database")
	}

	return nil
}
