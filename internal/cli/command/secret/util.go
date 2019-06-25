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

package secret

import "github.com/banzaicloud/banzai-cli/.gen/pipeline"

// filterOutHiddenSecrets filters out secrets with hidden tag
func filterOutHiddenSecrets(secrets []pipeline.SecretItem) []pipeline.SecretItem {
	return filterSecretsByTags(secrets, []string{"banzai:hidden"})
}

// filterOutHiddenAndReadonlySecrets filters out secrets with readonly and hidden tags
func filterOutHiddenAndReadonlySecrets(secrets []pipeline.SecretItem) []pipeline.SecretItem {
	return filterSecretsByTags(secrets, []string{"banzai:readonly", "banzai:hidden"})
}

// filterSecretsByTags filters out secrets by the given tags
func filterSecretsByTags(secrets []pipeline.SecretItem, tags []string) []pipeline.SecretItem {
	var filteredSecrets []pipeline.SecretItem
	for _, s := range secrets {
		if !contains(s.Tags, tags) {
			filteredSecrets = append(filteredSecrets, s)
		}
	}
	return filteredSecrets
}

func contains(slice []string, keys []string) bool {
	for _, item := range slice {
		for _, key := range keys {
			if item == key {
				return true
			}
		}
	}
	return false
}
