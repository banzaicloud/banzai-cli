## banzai cluster create

Create a cluster

### Synopsis

Create cluster based on json stdin or interactive session

```
banzai cluster create [flags]
```

### Options

```
  -f, --file string    Cluster descriptor file
  -h, --help           help for create
  -i, --interval int   Interval in seconds for polling cluster status (default 10)
  -w, --wait           Wait for cluster creation
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

* [banzai cluster](banzai_cluster.md)	 - Manage clusters

