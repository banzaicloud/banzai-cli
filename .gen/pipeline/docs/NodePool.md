# NodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Node pool name. | 
**Size** | **int32** | Node pool size. | 
**Labels** | **map[string]string** | Node pool labels. | [optional] 
**Autoscaling** | [**NodePoolAutoScaling**](NodePoolAutoScaling.md) |  | [optional] 
**InstanceType** | **string** | Machine instance type. | 
**Image** | **string** | Instance AMI. | [optional] 
**SpotPrice** | **string** | The upper limit price for the requested spot instance. If this field is left empty or 0 passed in on-demand instances used instead of spot instances. | [optional] 
**Subnet** | [**EksSubnet**](EKSSubnet.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


