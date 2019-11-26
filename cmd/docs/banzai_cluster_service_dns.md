## banzai cluster service dns

Manage cluster DNS service

### Synopsis

Manage cluster DNS service

```
banzai cluster service dns [flags]
```

### Options

```
      --cluster int32         ID of cluster to manage DNS cluster service of
      --cluster-name string   Name of cluster to manage DNS cluster service of
  -h, --help                  help for dns
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
* [banzai cluster service dns activate](banzai_cluster_service_dns_activate.md)	 - Activate the DNS service of a cluster
* [banzai cluster service dns deactivate](banzai_cluster_service_dns_deactivate.md)	 - Deactivate the DNS service of a cluster
* [banzai cluster service dns get](banzai_cluster_service_dns_get.md)	 - Get details of the DNS service for a cluster
* [banzai cluster service dns update](banzai_cluster_service_dns_update.md)	 - Update the DNS service of a cluster

