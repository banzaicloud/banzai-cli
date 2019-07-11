## banzai cluster shell

Start a shell or run a command with the cluster configured as kubectl context

### Synopsis

The banzai CLI's cluster shell command starts your default shell, or runs your specified program on your local machine within the Kubernetes context of your cluster. You can either run the command without arguments to interactively select a cluster, and get an interactive shell, select the cluster with the --cluster-name flag, or specify the command to run.

```
banzai cluster shell [command] [flags]
```

### Examples

```

			$ banzai cluster shell
			? Cluster: docs-example
			[docs-example]$ helm list
			...
			[docs-example]$ kubectl get nodes
			...
			[docs-example]$ exit
			INFO[0026] Command exited successfully

			$ banzai cluster shell --cluster-name docs-example kubectl get nodes
			INFO[0000] Running kubectl kubectl get nodes
			NAME                                    STATUS   ROLES    AGE   VERSION
			gke-docs-example-pool1-7a602b82-62w8    Ready    <none>   43m   v1.10.11-gke.1
			gke-docs-example-system-a16f163c-dvwj   Ready    <none>   43m   v1.10.11-gke.1
			INFO[0001] Command exited successfully
```

### Options

```
      --cluster int32         ID of cluster to run a shell for
      --cluster-name string   Name of cluster to run a shell for
  -h, --help                  help for shell
      --wrap-helm             Wrap the helm command with a version that downloads the matching version and creates a custom helm home (default true)
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

