# CreateNodePoolRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Node pool name. | 
**Size** | **int32** | Node pool size. | 
**Labels** | **map[string]string** | Node pool labels. | [optional] 
**Autoscaling** | [**NodePoolAutoScaling**](NodePoolAutoScaling.md) |  | [optional] 
**VolumeEncryption** | Pointer to [**EksNodePoolVolumeEncryption**](EKSNodePoolVolumeEncryption.md) |  | [optional] 
**VolumeSize** | **int32** | Size of the EBS volume in GBs of the nodes in the pool. | [optional] 
**VolumeType** | **string** | Type of the EBS volume of the nodes in the pool (default gp3). | [optional] 
**InstanceType** | **string** | Machine instance type. | 
**Image** | **string** | Instance AMI. | [optional] 
**SpotPrice** | **string** | The upper limit price for the requested spot instance. If this field is left empty or 0 passed in on-demand instances used instead of spot instances. | [optional] 
**SubnetId** | **string** |  | [optional] 
**SecurityGroups** | **[]string** | List of additional custom security groups for all nodes in the pool. | [optional] 
**UseInstanceStore** | **bool** | Setup available instance stores (NVMe disks) to use for Kubelet root if available. As a result emptyDir volumes will be provisioned on local instance storage disks. You can check out available instance storages here https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#instance-store-volumes. | [optional] 
**NodePools** | [**map[string]NodePool**](NodePool.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


