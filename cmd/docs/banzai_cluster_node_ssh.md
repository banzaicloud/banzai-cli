## banzai cluster node ssh

Connect to node with SSH

### Synopsis

Connect to node with SSH

```
banzai cluster node ssh [NODE_NAME] [flags]
```

### Options

```
      --cluster int32         ID of cluster to get
      --cluster-name string   Name of cluster to get
      --direct-connect        Use direct connection to the node internal or external IP (default)
  -h, --help                  help for ssh
      --namespace string      Namespace for the pod when using --pod-connect (default "pipeline-system")
      --node-name string      Node name
      --pod-connect           Create a pod on one of the nodes and connect to a node through that pod
      --ssh-port int          SSH port of the node to connect (default 22)
      --use-external-ip       Use internal IP of the node to connect
      --use-internal-ip       Use external IP of the node to connect (default)
      --use-node-affinity     Whether to use node affinity for pod scheduling when using --pod-connect
      --username string       Username
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

* [banzai cluster node](banzai_cluster_node.md)	 - Work with cluster nodes

