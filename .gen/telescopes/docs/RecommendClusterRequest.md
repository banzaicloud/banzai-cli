# RecommendClusterRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowBurst** | **bool** | Are burst instances allowed in recommendation | [optional] 
**AllowOlderGen** | **bool** | AllowOlderGen allow older generations of virtual machines (applies for EC2 only) | [optional] 
**Category** | **[]string** | Category specifies the virtual machine category | [optional] 
**Excludes** | **[]string** | Excludes is a blacklist - a slice with vm types to be excluded from the recommendation | [optional] 
**Includes** | **[]string** | Includes is a whitelist - a slice with vm types to be contained in the recommendation | [optional] 
**MaxNodes** | **int64** | Maximum number of nodes in the recommended cluster | [optional] 
**MinNodes** | **int64** | Minimum number of nodes in the recommended cluster | [optional] 
**NetworkPerf** | **[]string** | NetworkPerf specifies the network performance category | [optional] 
**OnDemandPct** | **int64** | Percentage of regular (on-demand) nodes in the recommended cluster | [optional] 
**SameSize** | **bool** | If true, recommended instance types will have a similar size | [optional] 
**SumCpu** | **float64** | Total number of CPUs requested for the cluster | [optional] 
**SumGpu** | **int64** | Total number of GPUs requested for the cluster | [optional] 
**SumMem** | **float64** | Total memory requested for the cluster (GB) | [optional] 
**Zone** | **string** | Availability zone that the cluster should expand to | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


