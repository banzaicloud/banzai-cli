# RecommendClusterScaleOutRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActualLayout** | [**[]NodePoolDesc**](NodePoolDesc.md) | Description of the current cluster layout in:body | [optional] 
**DesiredCpu** | **float64** | Total desired number of CPUs in the cluster after the scale out | [optional] 
**DesiredGpu** | **int64** | Total desired number of GPUs in the cluster after the scale out | [optional] 
**DesiredMem** | **float64** | Total desired memory (GB) in the cluster after the scale out | [optional] 
**Excludes** | **[]string** | Excludes is a blacklist - a slice with vm types to be excluded from the recommendation | [optional] 
**OnDemandPct** | **int64** | Percentage of regular (on-demand) nodes among the scale out nodes | [optional] 
**Zone** | **string** | Availability zone to be included in the recommendation | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


