## banzai secret update

Update secret

### Synopsis

Update an existing secret in Pipeline's secret store interactively, or based on a json request from stdin or a file

```
banzai secret update [flags]
```

### Examples

```

	Update secret
	---
	$ banzai secret update
	? Select secret: mysecret
	? Do you want modify fields of secret? Yes
	? Select field to modify: username
	? username myusername
	? Select field to modify: password
	? password mypassword
	? Select field to modify: skip
	? Do you want modify tags of secret? Yes
	? Do you want delete any tag of secret? Yes
	? Select tag(s) you want to delete: cli
	? Do you want to add tag(s) to this secret? Yes
	? Tag: banzai
	? Tag: skip
	? Do you want to validate this secret? Yes

	Update secret with flags
	---
	$ banzai secret update --name mysecret --validate false
	? Do you want modify fields of secret? Yes
	? Select field to modify: username
	? username myusername
	? Select field to modify: password
	? password mypassword
	? Select field to modify: skip
	? Do you want modify tags of secret? No
	
	Create secret via json
	---
	$ banzai secret update <<EOF
	> {
	>	"name": "mysecretname",
	>	"type": "password",
	>	"values": {
	>		"username": "myusername",
	>		"password": "mypassword"
	>	},
	>	"tags":[ "cli", "my-application" ],
	> 	"version": 1
	> }
	> EOF


```

### Options

```
  -f, --file string       Secret update descriptor file
  -h, --help              help for update
  -i, --id string         identification of the secret
  -n, --name string       Name of the secret
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

