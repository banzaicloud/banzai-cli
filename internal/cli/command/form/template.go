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
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/sprig"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type templateOptions struct {
	configFile string
	name       string
	force      bool
}

// NewTemplateCommand creates a new cobra.Command for `banzai form template`.
func NewTemplateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := templateOptions{}

	cmd := &cobra.Command{
		Use:        "template FORM_CONFIG [-n TEMPLATE_NAME] [--force]",
		Short:      "Execute form template(s)",
		Long:       "Execute form template(s) using values from the provided values in the config file",
		Deprecated: "This command will be removed later.",
		Args:       cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.configFile = args[0]
			runExecuteTemplate(banzaiCli, options)
		},
	}

	cmd.Flags().StringVarP(&options.name, "name", "n", "", "template name")
	cmd.Flags().BoolVar(&options.force, "force", false, "overwrite existing files")

	return cmd
}

func runExecuteTemplate(_ cli.Cli, options templateOptions) {
	file, err := readConfig(options.configFile)
	if err != nil {
		log.Fatal(err)
	}

	values := map[string]interface{}{}
	for _, group := range file.Form {
		for _, field := range group.Fields {
			values[field.Key] = field.Value
		}
	}

	var tmpls map[string]string
	if options.name != "" {
		for name, t := range file.Templates {
			if name == options.name {
				tmpls = map[string]string{name: t}
				break
			}
		}

		if len(tmpls) == 0 {
			log.Fatalf("could not find template with name: %s", options.name)
		}
	} else {
		tmpls = file.Templates
	}

	for filename, tmpl := range tmpls {
		t, err := template.New(options.name).Funcs(sprig.TxtFuncMap()).Option("missingkey=error").Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}

		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		filePath := path.Join(dir, filename)

		if !options.force {
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				overwrite := false
				prompt := &survey.Confirm{Message: fmt.Sprintf("%s already exists. Do you want to overwrite it?", filePath)}
				survey.AskOne(prompt, &overwrite)

				if !overwrite {
					continue
				}
			}
		}

		f, err := os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		err = t.Execute(f, values)
		if err != nil {
			log.Fatal(err)
		}
	}
}
