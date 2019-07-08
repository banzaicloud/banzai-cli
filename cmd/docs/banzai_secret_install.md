## banzai secret install

Install a secret to a cluster

### Synopsis

Install a particular secret to a cluster's namespace.

```
banzai secret install [flags]
```

### Examples

```

		Install secret
		-----
		$ banzai secret install --name mysecretname --cluster myClusterName <<EOF
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
  -c, --cluster string   Name or ID of the cluster to install the secret
  -f, --file string      Template descriptor file
  -h, --help             help for install
  -m, --merge            Set true to merge existing secret
  -n, --name string      Name of the secret to install
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

