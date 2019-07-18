## banzai bucket get

Get bucket

### Synopsis

Get bucket

```
banzai bucket get NAME [[--cloud=]CLOUD]] [flags]
```

### Options

```
      --cloud string             Cloud provider for the bucket
  -h, --help                     help for get
  -l, --location string          Location (e.g. us-central1) for the bucket
      --storage-account string   Storage account for the bucket (must be specified for Azure)
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

* [banzai bucket](banzai_bucket.md)	 - Manage buckets

