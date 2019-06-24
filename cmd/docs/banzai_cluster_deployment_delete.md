## banzai cluster deployment delete

Delete a deployment

### Synopsis

Delete a deployment identified by deployment release name.

```
banzai cluster deployment delete RELEASE-NAME [flags]
```

### Examples

```

			$ banzai cluster deployment delete test-deployment
			? Cluster  [Use arrows to move, type to filter]
			> pke-cluster-1

			Name  			 Status  Message            
			test-deployment  200     Deployment deleted!

			$ banzai cluster deployment delete test-deployment --cluster-name pke-cluster-1 --no-interactive
			Name  			 Status  Message            
			test-deployment  200     Deployment deleted!
```

### Options

```
      --cluster int32         ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -n, --cluster-name string   Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -h, --help                  help for delete
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

