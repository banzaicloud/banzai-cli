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

package cmd

import (
	"context"

	"github.com/banzaicloud/pipeline/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// secretCmd represents the secret command
var secretCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets", "s"},
	Short:   "List secrets",
	Run: func(cmd *cobra.Command, args []string) {
		pipeline := InitPipeline()
		orgId := GetOrgId(true)
		secrets, _, err := pipeline.SecretsApi.GetSecrets(context.Background(), orgId, &client.GetSecretsOpts{})
		if err != nil {
			log.Fatalf("could not list secrets: %v", err)
		}
		Out(secrets, []string{"Id", "Name", "Type", "UpdatedBy", "Tags"})
	},
}

func init() {
	rootCmd.AddCommand(secretCmd)
}
