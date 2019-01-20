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
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

const orgIdKey = "organization.id"

func searchOrganizationId(name string) int32 {
	pipeline := InitPipeline()
	orgs, _, err := pipeline.OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		cli.LogAPIError("list organizations", err, nil)
		return 0
	}
	for _, org := range orgs {
		if org.Name == name {
			return org.Id
		}
	}
	log.Errorf("Could not find organization %q", name)
	return 0
}

func GetOrgId(ask bool) int32 {
	id := viper.GetInt32(orgIdKey)
	if id != 0 {
		return id
	}
	if ask && !isInteractive() {
		log.Fatal("No organization is selected. Use the --organization switch, or set it using `banzai org`.")
	}
	pipeline := InitPipeline()
	orgs, _, err := pipeline.OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		log.Fatalf("could not list organizations: %v", err)
	}
	orgSlice := make([]string, len(orgs))
	for i, org := range orgs {
		orgSlice[i] = org.Name
	}
	name := ""
	survey.AskOne(&survey.Select{Message: "Organization:", Options: orgSlice}, &name, survey.Required)
	id = searchOrganizationId(name)
	if id != 0 {
		viper.Set(orgIdKey, id)
	}
	return id
}
