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

package process

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/pkg/spinner"
)

func TailProcess(banzaiCli cli.Cli, processId string) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	statuses := map[string]*spinner.Status{}

	for {
		process, resp, err := client.ProcessesApi.GetProcess(context.Background(), orgID, processId)
		if err != nil {
			return errors.WrapIf(err, "failed to list node pool update processes")
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.NewWithDetails("node pool update process list failed with http status code", "status_code", resp.StatusCode)
		}

		for i, event := range process.Events {
			if s, ok := statuses[event.Type]; !ok {
				status := spinner.NewStatus()
				status.Start(fmt.Sprintf("[%s] executing %s activity %s", event.Timestamp.Local().Format(time.RFC3339), event.Type, event.Log))
				statuses[event.Type] = status

				if i == len(process.Events)-1 {
					time.Sleep(2 * time.Second)
				}
			} else if event.Status != pipeline.RUNNING {
				s.End(event.Status == pipeline.FINISHED)
			} else {
				if i == len(process.Events)-1 {
					time.Sleep(2 * time.Second)
				}
			}
		}

		if process.Status == pipeline.FINISHED {
			return nil
		} else if process.Status == pipeline.FAILED {
			return errors.NewWithDetails("process failed", "err", process.Log)
		}
	}
}
