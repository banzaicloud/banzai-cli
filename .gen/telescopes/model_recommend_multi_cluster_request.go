/*
 * Cluster Recommender.
 *
 * This project can be used to recommend instance type groups on different cloud providers consisting of regular and spot/preemptible instances. The main goal is to provide and continuously manage a cost-effective but still stable cluster layout that's built up from a diverse set of regular and spot instances.
 *
 * API version: 0.5.1
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package telescopes

// MultiClusterRecommendationReq encapsulates the recommendation input data
type RecommendMultiClusterRequest struct {
	// Are burst instances allowed in recommendation
	AllowBurst bool `json:"allowBurst,omitempty"`
	// AllowOlderGen allow older generations of virtual machines (applies for EC2 only)
	AllowOlderGen bool `json:"allowOlderGen,omitempty"`
	// Category specifies the virtual machine category
	Category []string `json:"category,omitempty"`
	Continents []string `json:"continents,omitempty"`
	// Excludes is a blacklist - a slice with vm types to be excluded from the recommendation
	Excludes map[string]map[string][]string `json:"excludes,omitempty"`
	// Includes is a whitelist - a slice with vm types to be contained in the recommendation
	Includes map[string]map[string][]string `json:"includes,omitempty"`
	// Maximum number of nodes in the recommended cluster
	MaxNodes int64 `json:"maxNodes,omitempty"`
	// Minimum number of nodes in the recommended cluster
	MinNodes int64 `json:"minNodes,omitempty"`
	// NetworkPerf specifies the network performance category
	NetworkPerf []string `json:"networkPerf,omitempty"`
	// Percentage of regular (on-demand) nodes in the recommended cluster
	OnDemandPct int64 `json:"onDemandPct,omitempty"`
	Providers []Provider `json:"providers,omitempty"`
	// Maximum number of response per service
	RespPerService int64 `json:"respPerService,omitempty"`
	// If true, recommended instance types will have a similar size
	SameSize bool `json:"sameSize,omitempty"`
	// Total number of CPUs requested for the cluster
	SumCpu float64 `json:"sumCpu,omitempty"`
	// Total number of GPUs requested for the cluster
	SumGpu int64 `json:"sumGpu,omitempty"`
	// Total memory requested for the cluster (GB)
	SumMem float64 `json:"sumMem,omitempty"`
}