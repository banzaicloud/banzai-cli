# EksNodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InstanceType** | **string** |  | 
**SpotPrice** | **string** |  | 
**Autoscaling** | **bool** |  | [optional] 
**Count** | **int32** |  | [optional] 
**MinCount** | **int32** |  | 
**MaxCount** | **int32** |  | 
**Labels** | **map[string]string** |  | [optional] 
**VolumeEncryption** | Pointer to [**EksNodePoolVolumeEncryption**](EKSNodePoolVolumeEncryption.md) |  | [optional] 
**VolumeSize** | **int32** | Size of the EBS volume in GBs of the nodes in the pool. | [optional] 
**Image** | **string** |  | [optional] 
**Subnet** | [**EksSubnet**](EKSSubnet.md) |  | [optional] 
**SecurityGroups** | **[]string** | List of additional custom security groups for all nodes in the pool. | [optional] 
**UseInstanceStore** | **bool** | Setup available instance stores (NVMe disks) to use for Kubelet root if available. As a result emptyDir volumes will be provisioned on local instance storage disks. You can check out available instance storages here https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#instance-store-volumes. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


