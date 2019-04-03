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

package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/pkg/formatting"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

var rootOptions struct {
	CfgFile string
	Output  string
}

func InitPipeline() *pipeline.APIClient {
	config := pipeline.NewConfiguration()
	config.BasePath = viper.GetString("pipeline.basepath")
	config.UserAgent = "banzai-cli/1.0.0/go"
	config.HTTPClient = oauth2.NewClient(nil, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("pipeline.token")},
	))

	return pipeline.NewAPIClient(config)
}

func Out1(data interface{}, fields []string) {
	Out([]interface{}{data}, fields)
}

func Out(data interface{}, fields []string) {
	switch rootOptions.Output {
	case "json":
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatalf("can't marshal output: %v", err)
		}
		fmt.Printf("%s\n", bytes)

	case "yaml":
		bytes, err := yaml.Marshal(data)
		if err != nil {
			log.Fatalf("can't marshal output: %v", err)
		}
		fmt.Printf("%s\n", bytes)

	default:
		table := formatting.NewTable(data, fields)
		out := table.Format(isColor())
		fmt.Println(out)
	}
}

func isInteractive() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
		return !viper.GetBool("formatting.no-interactive")
	}
	return viper.GetBool("formatting.force-interactive")
}

func isColor() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return !viper.GetBool("formatting.no-color")
	}
	return viper.GetBool("formatting.force-color")
}
