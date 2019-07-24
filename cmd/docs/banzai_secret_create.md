## banzai secret create

Create secret

### Synopsis

Create a secret in Pipeline's secret store interactively, or based on a json request from stdin or a file

```
banzai secret create [flags]
```

### Examples

```

	Create secret
	---
	$ banzai secret create
	? Secret name mysecretname
	? Choose secret type: password
	? Set 'username' field: myusername
	? Set 'password' field: mypassword
	? Do you want to add tag(s) to this secret? Yes
	? Tag: tag1
	? Tag: tag2
	? Tag: skip

	Create secret with flags
	---
	$ banzai secret create --name mysecretname --type password --tag=cli --tag=my-application
	? Set 'username' field: myusername
	? Set 'password' field: mypassword

	Create secret via json
	---
	$ banzai secret create <<EOF
	> {
	>	"name": "mysecretname",
	>	"type": "password",
	>	"values": {
	>		"username": "myusername",
	>		"password": "mypassword"
	>	},
	>	"tags":[ "cli", "my-application" ]
	> }
	> EOF
		
```

### Options

```
  -f, --file string       Secret creation descriptor file
  -h, --help              help for create
      --magic             Try to import credentials from local environment (AWS only for now)
  -n, --name string       Name of the secret
      --tag stringArray   Tags to add to the secret
  -t, --type string       Type of the secret
  -v, --validate string   Secret validation (true|false)
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

