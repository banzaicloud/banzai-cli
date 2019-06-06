/*
 * Pipeline API
 *
 * Pipeline v0.3.0 swagger
 *
 * API version: 0.21.2
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pipeline

type NodeItemStatus struct {
	Capacity NodeItemStatusCapacity `json:"capacity,omitempty"`
	Allocatable NodeItemStatusAllocatable `json:"allocatable,omitempty"`
	Conditions []NodeItemStatusConditions `json:"conditions,omitempty"`
	Addresses []NodeItemStatusAddresses `json:"addresses,omitempty"`
	DaemonEndpoints NodeItemStatusDaemonEndpoints `json:"daemonEndpoints,omitempty"`
	NodeInfo NodeItemStatusNodeInfo `json:"nodeInfo,omitempty"`
	Images []NodeItemStatusImages `json:"images,omitempty"`
}
