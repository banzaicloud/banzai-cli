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

package nodepool

import (
	"context"
	"log"
	"sort"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/spf13/cobra"
)

const (
	nodePoolAutoscalingDisabled = nodePoolAutoscaling("Disabled")

	nodePoolAutoscalingEnabled = nodePoolAutoscaling("Enabled")
)

type nodePoolAutoscaling string

func newNodePoolAutoscaling(isEnabled bool) (autoscaling nodePoolAutoscaling) {
	if isEnabled {
		return nodePoolAutoscalingEnabled
	}

	return nodePoolAutoscalingDisabled
}

type nodePoolListItem struct {
	Name             string
	Size             int32
	Autoscaling      nodePoolAutoscaling
	MinimumSize      int32
	MaximumSize      int32
	VolumeEncryption nodePoolVolumeEncryption
	VolumeSize       int32
	VolumeType       string
	InstanceType     string
	Image            string
	SpotPrice        string
	SubnetID         string
	SecurityGroups   []string
	Status           string
	StatusMessage    string
}

type nodePoolListOptions struct {
	clustercontext.Context
}

type nodePoolVolumeEncryption string

const (
	// nodePoolVolumeEncryptionAWSAccountDefault indicates the default AWS
	// account volume encryption setting is being used.
	nodePoolVolumeEncryptionAWSAccountDefault nodePoolVolumeEncryption = "AWS account default"

	// nodePoolVolumeEncryptionDisabled indicates no volume encryption.
	nodePoolVolumeEncryptionDisabled nodePoolVolumeEncryption = "Disabled"

	// nodePoolVolumeEncryptionEnabled indicates a volume encryption with the
	// AWS account default volume encryption key for the account.
	nodePoolVolumeEncryptionEnabled nodePoolVolumeEncryption = "Enabled (AWS default key)"
)

// newNodePoolVolumeEncryptionString transforms the volume encryption object
// into an explanatory string.
func newNodePoolVolumeEncryption(volumeEncryption *pipeline.EksNodePoolVolumeEncryption) nodePoolVolumeEncryption {
	if volumeEncryption == nil {
		return nodePoolVolumeEncryptionAWSAccountDefault
	} else if volumeEncryption.Enabled == false {
		return nodePoolVolumeEncryptionDisabled
	} else if volumeEncryption.EncryptionKeyARN == "" {
		return nodePoolVolumeEncryptionEnabled
	}

	return nodePoolVolumeEncryption(volumeEncryption.EncryptionKeyARN)
}

func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := nodePoolListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List node pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNodePoolList(banzaiCli, options)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "nodepool-list")

	return cmd
}

func runNodePoolList(banzaiCli cli.Cli, options nodePoolListOptions) error {
	pipelineClient := banzaiCli.Client()
	organizationID := banzaiCli.Context().OrganizationID()

	if err := options.Init(); err != nil {
		return err
	}

	clusterID := options.ClusterID()

	nodePools, _, err := pipelineClient.ClustersApi.ListNodePools(context.Background(), organizationID, clusterID)
	if err != nil {
		cli.LogAPIError("list node pools", err, clusterID)
		log.Fatalf("could not list node pools: %v", err)
	}

	nodePoolListItems := make([]nodePoolListItem, len(nodePools))
	for nodePoolIndex, nodePool := range nodePools {
		nodePoolListItems[nodePoolIndex] = nodePoolListItem{
			Name:             nodePool.Name,
			Size:             nodePool.Size,
			Autoscaling:      newNodePoolAutoscaling(nodePool.Autoscaling.Enabled),
			MinimumSize:      nodePool.Autoscaling.MinSize,
			MaximumSize:      nodePool.Autoscaling.MaxSize,
			VolumeEncryption: newNodePoolVolumeEncryption(nodePool.VolumeEncryption),
			VolumeSize:       nodePool.VolumeSize,
			VolumeType:       nodePool.VolumeType,
			InstanceType:     nodePool.InstanceType,
			Image:            nodePool.Image,
			SpotPrice:        nodePool.SpotPrice,
			SubnetID:         nodePool.SubnetId,
			SecurityGroups:   nodePool.SecurityGroups,
			Status:           nodePool.Status,
			StatusMessage:    nodePool.StatusMessage,
		}
	}

	sort.Slice(nodePoolListItems, func(firstIndex, secondIndex int) (isLessThan bool) {
		return nodePoolListItems[firstIndex].Name < nodePoolListItems[secondIndex].Name
	})

	format.NodePoolsWrite(banzaiCli, nodePoolListItems)

	return nil
}
