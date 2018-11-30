This is a command line interface under heavy development for the Banzai Cloud platform.

### Installation

```
$ go get github.com/banzaicloud/banzai-cli/cmd/banzai
```

### Use

```
A command line client for the Banzai Pipeline platform.

Usage:
  banzai [command]

Available Commands:

  login        Configure and log in to a Banzai Cloud context

  cluster      Handle clusters
    create      Create cluster based on json stdin or interactive session
    delete      Delete a cluster
    get         Get cluster details
    list        List clusters
    shell       Start a shell or run a command with the cluster configured as kubectl context

  organization List and select organizations
    list        List organizations
    select      Select organization

  secret       List secrets
  help         Help about any command

Flags:
      --color                use colors on non-tty outputs
      --config string        config file (default is $HOME/.banzai/config.yaml)
  -h, --help                 help for banzai
      --interactive          ask questions interactively even if stdin or stdout is non-tty
      --no-color             never display color output
      --no-interactive       never ask questions interactively
      --organization int32   organization id
  -o, --output string        output format (default|yaml|json) (default "default")
      --verbose              more verbose output

Use "banzai [command] --help" for more information about a command.
```
