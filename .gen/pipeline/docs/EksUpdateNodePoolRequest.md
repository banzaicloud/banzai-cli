# EksUpdateNodePoolRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | **int32** | Node pool size. | 
**Labels** | **map[string]string** | Node pool labels. | [optional] 
**Autoscaling** | [**NodePoolAutoScaling**](NodePoolAutoScaling.md) |  | [optional] 
**VolumeSize** | **int32** | Size of the EBS volume in GBs of the nodes in the pool. | [optional] 
**InstanceType** | **string** | The instance type to use for your node pool. | [optional] 
**Image** | **string** | The instance AMI to use for your node pool. | [optional] 
**Version** | **string** | The Kubernetes version to use for your node pool. | [optional] 
**SpotPrice** | **string** | The upper limit price for the requested spot instance. If this field is empty or 0 on-demand instances are used instead of spot instances. | [optional] 
**SecurityGroups** | Pointer to **[]string** | List of additional custom security groups for all nodes in the pool. | [optional] 
**Options** | [**BaseUpdateNodePoolOptions**](BaseUpdateNodePoolOptions.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


