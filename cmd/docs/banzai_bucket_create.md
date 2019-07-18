## banzai bucket create

Create bucket

### Synopsis

Create object storage bucket on supported cloud providers

```
banzai bucket create NAME [[--cloud=]CLOUD] [[--location=]LOCATION] [[--secret-id=]SECRET_ID] [flags]
```

### Options

```
  -c, --cloud string             Cloud provider for the bucket
  -h, --help                     help for create
  -l, --location string          Location for the bucket
      --resource-group string    Resource group for the bucket (must be specified for Azure)
  -s, --secret-id string         Secret ID of the used secret to create the bucket
      --storage-account string   Storage account for the bucket (must be specified for Azure)
  -w, --wait                     Wait for bucket creation
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

