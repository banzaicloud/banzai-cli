## banzai deployment create

Creates a deployment

### Synopsis

Creates a deployment based on deployment descriptor JSON read from stdin or file.

```
banzai deployment create [flags]
```

### Examples

```

        # Create deployment from file using interactive mode
        ----------------------------------------------------
        $ banzai deployment create
        ? Cluster  [Use arrows to move, type to filter]
        > pke-cluster-1
        ? Load a JSON or YAML file: [? for help] /var/tmp/wordpress.json

        ReleaseName       Notes
        torpid-armadillo  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK


        # Create deployment from stdin
        ------------------------------
        $ banzai deployment create --cluster-name pke-cluster-1 -f -<<EOF
        > {
        >   "name": "stable/wordpress",
        >   "releasename": "",
        >   "namespace": "default",
        >   "version": "5.12.4",
        >   "dryRun": false,
        >   "values": {
        >		"replicaCount": 2
        >   }
        > }
        > EOF

        ReleaseName       Notes
        lumbering-lizard  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK  

        $ echo '{"name":"stable/wordpress","releasename":"my-wordpress-1"}' |  banzai deployment create --cluster-name pke-cluster-1
        ReleaseName     Notes
        my-wordpress-1  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK

        # Create deployment from file
        -----------------------------
        $ banzai deployment create --cluster-name pke-cluster-1 --file /var/tmp/wordpress.json --no-interactive

        ReleaseName         Notes
        eyewitness-opossum  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK
```

### Options

```
      --cluster int32         ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -n, --cluster-name string   Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster
  -f, --file string           Deployment descriptor file
  -h, --help                  help for create
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

* [banzai deployment](banzai_deployment.md)	 - Manage deployments

