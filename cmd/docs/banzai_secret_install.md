## banzai secret install

Install a secret to a cluster

### Synopsis

Install a particular secret from Pipeline as a Kubernetes secret to a cluster.

```
banzai secret install [flags]
```

### Examples

```

		Install secret
		-----
		$ banzai secret install --name mysecretname --cluster-name myClusterName <<EOF
		> {
		> 	"namespace": "default",
		> 	"spec": {
		> 		"ROOT_USER": {
		> 			"source": "AWS_ACCESS_KEY_ID"
		> 		}
		> 	}
		> }
		> EOF
		
```

### Options

```
      --cluster int32         ID of cluster to install secret on
      --cluster-name string   Name of cluster to install secret on
  -f, --file string           Template descriptor file
  -h, --help                  help for install
  -m, --merge                 Merge fields to an existing Kubernetes secret
  -n, --name string           Name of the Pipeline secret to use
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

* [banzai secret](banzai_secret.md)	 - Manage secrets

