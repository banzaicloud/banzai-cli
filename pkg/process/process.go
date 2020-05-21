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
	"github.com/emirpasic/gods/maps/linkedhashmap"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/pkg/spinner"
)

const processVisibleThreshold = 3

type ProcessFailedError struct {
	msg string
}

func (e ProcessFailedError) Error() string {
	return e.msg
}

func IsProcessFailedError(err error) bool {
	_, ok := err.(ProcessFailedError)
	return ok
}

func TailProcess(banzaiCli cli.Cli, processId string) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	status := spinner.NewStatus()
	status.Start(fmt.Sprintf("[%s] tailing process %s", time.Now().Local().Format(time.RFC3339), processId))

	statuses := map[string]*spinner.Status{}

	processVisibleChecks := 0

	eventsToBeProcessed := linkedhashmap.New()
	processedEvents := map[int32]pipeline.ProcessEvent{}

	for {
		process, resp, err := client.ProcessesApi.GetProcess(context.Background(), orgID, processId)
		// we need to give some time for the process to appear
		if resp != nil && resp.StatusCode == 404 && processVisibleChecks < processVisibleThreshold {
			processVisibleChecks++
			time.Sleep(2 * time.Second)
			continue
		}
		if err != nil {
			return errors.WrapIf(err, "failed to list node pool update processes")
		}
		defer resp.Body.Close()
		status.End(true)

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return errors.NewWithDetails("node pool update process list failed with http status code", "status_code", resp.StatusCode)
		}

		for _, e := range process.Events {
			if _, ok := processedEvents[e.Id]; !ok {
				eventsToBeProcessed.Put(e.Id, e)
			}
		}

		for i := 0; i < len(eventsToBeProcessed.Keys()); {
			key := eventsToBeProcessed.Keys()[i]
			value, _ := eventsToBeProcessed.Get(key)
			event := value.(pipeline.ProcessEvent)

			if s, ok := statuses[event.Type]; !ok {
				// TODO(nandi) don't let multiple statuses run for now for different event types
				if len(statuses) > 0 {
					i++
					continue
				}

				status := spinner.NewStatus()
				status.Start(fmt.Sprintf("[%s] executing %s activity %s", event.Timestamp.Local().Format(time.RFC3339), event.Type, event.Log))
				statuses[event.Type] = status

				processedEvents[event.Id] = event
				eventsToBeProcessed.Remove(key)
			} else if event.Status != pipeline.RUNNING {
				s.End(event.Status == pipeline.FINISHED)
				delete(statuses, event.Type)

				processedEvents[event.Id] = event
				eventsToBeProcessed.Remove(key)

				// let's go back to the beginning of the stream and check for unprocessed events
				i = 0
			} else {
				// TODO(nandi) don't let multiple statuses run for now for same event types
				i++
			}
		}

		if process.Status == pipeline.FINISHED {
			_, _ = fmt.Fprintf(banzaiCli.Out(), "%s process finished", process.Type)
			return nil
		} else if process.Status != pipeline.RUNNING {
			return ProcessFailedError{fmt.Sprintf("%s process %s: %s", process.Type, process.Status, process.Log)}
		}

		time.Sleep(2 * time.Second)
	}
}
