## banzai cluster service vault

Manage cluster Vault service

### Synopsis

Manage cluster Vault service

```
banzai cluster service vault [flags]
```

### Options

```
      --cluster int32         ID of cluster to manage Vault cluster service of
      --cluster-name string   Name of cluster to manage Vault cluster service of
  -h, --help                  help for vault
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

* [banzai cluster service](banzai_cluster_service.md)	 - Manage cluster integrated services
* [banzai cluster service vault activate](banzai_cluster_service_vault_activate.md)	 - Activate the Vault service of a cluster
* [banzai cluster service vault deactivate](banzai_cluster_service_vault_deactivate.md)	 - Deactivate the Vault service of a cluster
* [banzai cluster service vault get](banzai_cluster_service_vault_get.md)	 - Get details of the Vault service for a cluster
* [banzai cluster service vault update](banzai_cluster_service_vault_update.md)	 - Update the Vault service of a cluster

