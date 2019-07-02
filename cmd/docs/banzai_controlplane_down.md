## banzai controlplane down

Destroy the controlplane

### Synopsis

Destroy a controlplane based on json stdin or interactive session

```
banzai controlplane down [flags]
```

### Options

```
  -h, --help               help for down
      --image-pull         Pull cp-installer image even if it's present locally (default true)
      --image-tag string   Tag of banzaicloud/cp-installer Docker image to use (default "latest")
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

* [banzai controlplane](banzai_controlplane.md)	 - Manage controlplane

