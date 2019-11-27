## banzai cluster nodepool delete

Delete a node pool for a given cluster

### Synopsis

Delete a node pool for a given cluster

```
banzai cluster nodepool delete [NODE_POOL_NAME] [flags]
```

### Options

```
      --cluster int32           ID of cluster to delete
      --cluster-name string     Name of cluster to delete
  -h, --help                    help for delete
      --node-pool-name string   Node pool name
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

* [banzai cluster nodepool](banzai_cluster_nodepool.md)	 - Work with cluster nodepools

