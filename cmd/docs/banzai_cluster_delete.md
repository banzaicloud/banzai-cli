## banzai cluster delete

Delete a cluster

### Synopsis

Delete a cluster. The cluster to delete is identified either by its name or the numerical ID. In case of interactive mode banzai CLI will prompt for a confirmation.

```
banzai cluster delete [--cluster=ID | [--cluster-name=]NAME] [flags]
```

### Options

```
      --cluster int32         ID of cluster to delete
      --cluster-name string   Name of cluster to delete
  -f, --force                 Allow non-graceful cluster deletion
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

* [banzai cluster](banzai_cluster.md)	 - Manage clusters

