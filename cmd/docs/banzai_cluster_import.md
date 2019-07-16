## banzai cluster import

Import an existing cluster (EXPERIMENTAL)

### Synopsis

This is an experimental feature. You can import an existing Kubernetes cluster into Pipeline. Some Pipeline features may not work as expected.

```
banzai cluster import [flags]
```

### Examples

```
banzai cluster import --name myimportedcluster --kubeconfig=kube.conf
kubectl config view --minify --raw | banzai cluster import -n myimportedcluster
```

### Options

```
  -h, --help                help for import
      --kubeconfig string   Kubeconfig file (with embed client cert/key for the user entry)
  -n, --name string         Name of the cluster
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

