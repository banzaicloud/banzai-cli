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

// ClusterRecommendationResp encapsulates recommendation result data
type ClusterRecommendationResp struct {
	Accuracy ClusterRecommendationAccuracy `json:"accuracy,omitempty"`
	// Recommended node pools
	NodePools []NodePool `json:"nodePools,omitempty"`
	// The cloud provider
	Provider string `json:"provider,omitempty"`
	// Service's region
	Region string `json:"region,omitempty"`
	// Provider's service
	Service string `json:"service,omitempty"`
	// Availability zone in the recommendation - a multi-zone recommendation means that all node pools should expand to all zones
	Zone string `json:"zone,omitempty"`
}