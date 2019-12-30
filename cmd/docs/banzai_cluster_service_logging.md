## banzai cluster service logging

Manage cluster Logging service

### Synopsis

Manage cluster Logging service

```
banzai cluster service logging [flags]
```

### Options

```
      --cluster int32         ID of cluster to manage Logging cluster service of
      --cluster-name string   Name of cluster to manage Logging cluster service of
  -h, --help                  help for logging
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
* [banzai cluster service logging activate](banzai_cluster_service_logging_activate.md)	 - Activate the Logging service of a cluster
* [banzai cluster service logging deactivate](banzai_cluster_service_logging_deactivate.md)	 - Deactivate the Logging service of a cluster
* [banzai cluster service logging get](banzai_cluster_service_logging_get.md)	 - Get details of the Logging service for a cluster
* [banzai cluster service logging update](banzai_cluster_service_logging_update.md)	 - Update the Logging service of a cluster

