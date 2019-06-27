## banzai deployment get

Get deployment details

### Synopsis

Get the details of a deployment identified by deployment release name. In order to display deployment current values and notes use --output=(json|yaml)

```
banzai deployment get RELEASE-NAME [flags]
```

### Examples

```

			$ banzai deployment get dns
			? Cluster  [Use arrows to move, type to filter]
			> pke-cluster-1
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2 

			$ banzai deployment get dns --cluster-name pke-cluster-1
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2

			$ banzai deployment get dns --cluster 1846
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2
```

### Options

```
      --cluster int32         ID of the cluster to get deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -n, --cluster-name string   Name of the cluster to get deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -h, --help                  help for get
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

* [banzai deployment](banzai_deployment.md)	 - Manage deployments

