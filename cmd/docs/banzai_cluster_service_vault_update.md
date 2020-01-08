## banzai cluster service vault update

Update the Vault service of a cluster

### Synopsis

Update the Vault service of a cluster

```
banzai cluster service vault update [flags]
```

### Options

```
      --cluster int32         ID of cluster to update Vault cluster service for
      --cluster-name string   Name of cluster to update Vault cluster service for
  -f, --file string           Service specification file
  -h, --help                  help for update
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

* [banzai cluster service vault](banzai_cluster_service_vault.md)	 - Manage cluster Vault service

