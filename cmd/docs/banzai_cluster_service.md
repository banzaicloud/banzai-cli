## banzai cluster service

Manage cluster integrated services

### Synopsis

Manage cluster integrated services

```
banzai cluster service [flags]
```

### Options

```
      --cluster int32         ID of cluster to list services
      --cluster-name string   Name of cluster to list services
  -h, --help                  help for service
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
* [banzai cluster service dns](banzai_cluster_service_dns.md)	 - Manage cluster DNS service
* [banzai cluster service list](banzai_cluster_service_list.md)	 - List active (and pending) integrated services of a cluster
* [banzai cluster service logging](banzai_cluster_service_logging.md)	 - Manage cluster Logging service
* [banzai cluster service monitoring](banzai_cluster_service_monitoring.md)	 - Manage cluster Monitoring service
* [banzai cluster service securityscan](banzai_cluster_service_securityscan.md)	 - Manage cluster securityscan service
* [banzai cluster service vault](banzai_cluster_service_vault.md)	 - Manage cluster Vault service

