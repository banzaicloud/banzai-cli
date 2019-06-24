## banzai cluster deployment list

List deployments

### Synopsis

List deployments

```
banzai cluster deployment list [flags]
```

### Examples

```

				$ banzai cluster deployment ls

				? Cluster  [Use arrows to move, type to filter]
				> pke-cluster-1

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3      

				$ banzai cluster deployment ls --cluster-name pke-cluster-1

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3

				$ banzai cluster deployment ls --cluster 1846

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3
```

### Options

```
      --cluster int32         ID of the cluster which to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -n, --cluster-name string   Name of the cluster which to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -h, --help                  help for list
```

### Options inherited from parent commands

```
      --color                use colors on non-tty outputs
      --config string        config file (default is $BANZAICONFIG or $HOME/.banzai/config.yaml)
      --interactive          ask questions interactively even if stdin or stdout is non-tty
      --no-color             never display color output
      --no-interactive       never ask questions interactively
      --organization int32   organization id
  -o, --output string        output format (default|yaml|json) (default "default")
      --verbose              more verbose output
```

### SEE ALSO

* [banzai cluster deployment](banzai_cluster_deployment.md)	 - Manage deployments

