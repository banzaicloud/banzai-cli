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

package input

import (
	"context"
	"log"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"gopkg.in/AlecAivazis/survey.v1"
)

// AskOrganization asks for an organization.
func AskOrganization(banzaiCli cli.Cli) int32 {
	if err := checkPipeline(banzaiCli); err != nil {
		log.Fatal(err)
	}
	orgs, _, err := banzaiCli.Client().OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		log.Fatalf("could not list organizations: %v", err)
	}

	if len(orgs) == 1 {
		log.Printf("selecting organization %q", orgs[0].Name)
		return orgs[0].Id
	}

	orgSelection := make([]string, len(orgs))
	orgResultMap := make(map[string]int32, len(orgs))
	for i, org := range orgs {
		orgSelection[i] = org.Name
		orgResultMap[org.Name] = org.Id
	}

	var name string
	err = survey.AskOne(&survey.Select{Message: "Organization:", Options: orgSelection}, &name, survey.Required)
	if err != nil {
		log.Printf("could not choose an organization: %v", err)
	}

	return orgResultMap[name]
}

// GetOrganization returns the current organization.
func GetOrganization(banzaiCli cli.Cli) int32 {
	id := banzaiCli.Context().OrganizationID()

	if id == 0 {
		id = AskOrganization(banzaiCli)
	}

	return id
}

// GetOrganizations returns a map with the list of organizations
// where the key is the organization name and the value is the id
func GetOrganizations(banzaiCli cli.Cli) (map[string]int32, error) {
	orgs, _, err := banzaiCli.Client().OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		return nil, emperror.Wrap(err, "could not list organizations")
	}

	orgMap := make(map[string]int32, len(orgs))
	for _, org := range orgs {
		orgMap[org.Name] = org.Id
	}

	return orgMap, nil
}
