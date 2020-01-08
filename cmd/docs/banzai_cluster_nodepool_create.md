## banzai cluster nodepool create

Create a new node pool

### Synopsis

Create a new node pool

```
banzai cluster nodepool create [flags]
```

### Options

```
      --cluster int32         ID of cluster to create
      --cluster-name string   Name of cluster to create
  -f, --file string           Node pool descriptor file
  -h, --help                  help for create
  -n, --name string           Node pool name
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

* [banzai cluster nodepool](banzai_cluster_nodepool.md)	 - Manage node pools

