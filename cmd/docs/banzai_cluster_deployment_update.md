## banzai cluster deployment update

Updates a deployment

### Synopsis

Updates a deployment identified by release name using a deployment descriptor JSON read from stdin or file.

```
banzai cluster deployment update [flags]
```

### Examples

```

		# Update deployment from file using interactive mode
        ----------------------------------------------------
        $ banzai cluster deployment update
        ? Cluster pke-cluster-1
        ? Release name  [Use arrows to move, type to filter]
        > hazelcast-1
        exacerbated-narwhal
        luminous-hare

        ? Load a JSON or YAML file: [? for help] /var/tmp/hazelcast.json

        ReleaseName  Notes
        hazelcast-1  aGF6ZWxjYXN0LTEgcmVsZWFzZQo=

        # Update deployment from stdin
        ------------------------------
        $ banzai cluster deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1 -f -<<EOF
        > {
        >     "name": "stable/hazelcast",
        >     "version": "1.3.3",
        >     "reuseValues": true,
        >     "values": {
        >         "cluster": {
        >             "memberCount": 5
        >         }
        >     } 
        > }
        > EOF

        $ echo '{"name":"stable/hazelcast","version":"1.3.3","reuseValues":true,"values":{"cluster":{"memberCount":5}}}' | banzai cluster deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1

        # Update deployment from file
        -----------------------------
        $ banzai cluster deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1 -f /var/tmp/hazelcast.json
```

### Options

```
      --cluster int32         ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -n, --cluster-name string   Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -f, --file string           Deployment descriptor file
  -h, --help                  help for update
  -r, --release-name string   Deployment release name
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

* [banzai cluster deployment](banzai_cluster_deployment.md)	 - Manage deployments

