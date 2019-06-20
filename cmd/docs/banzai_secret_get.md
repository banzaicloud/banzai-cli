## banzai secret get

Get a secret

### Synopsis

Get a secret

```
banzai secret get ([--name=]NAME | --id=ID) [flags]
```

### Options

```
  -h, --help          help for get
  -H, --hide          Hide secret contents in the output
  -i, --id string     ID of secret to get
  -n, --name string   Name of secret to get
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

