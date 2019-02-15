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
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type templateOptions struct {
	configFile string
	name       string
	directory  string
}

// NewTemplateCommand creates a new cobra.Command for `banzai form template`.
func NewTemplateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := templateOptions{}

	cmd := &cobra.Command{
		Use:   "template CONFIG [-n NAME] [-d DIRECTORY]",
		Short: "Execute form template",
		Long:  "Execute form template(s) using values from the provided values in the config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.configFile = args[0]
			if options.directory != "" {
				path, err := filepath.Abs(options.directory)
				if err != nil {
					log.Fatal(err)
				}

				options.directory = path
			}

			runExecuteTemplate(banzaiCli, options)
		},
	}

	cmd.Flags().StringVarP(&options.directory, "directory", "d", "", "write executed template files to this directory")
	cmd.Flags().StringVarP(&options.name, "name", "n", "", "template name")

	return cmd
}

func runExecuteTemplate(banzaiCli cli.Cli, options templateOptions) {
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

	// Create directory if it does not exist
	if options.directory != "" {
		err = os.MkdirAll(options.directory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	for filename, tmpl := range tmpls {
		t, err := template.New(options.name).Funcs(sprig.TxtFuncMap()).Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}

		var f io.Writer
		if options.directory != "" {
			f, err = os.Create(path.Join(options.directory, filename))
			if err != nil {
				log.Fatal(err)
			}

			defer f.(io.WriteCloser).Close()
		} else {
			f = os.Stdout
		}

		err = t.Execute(f, values)
		if err != nil {
			log.Fatal(err)
		}
	}
}
