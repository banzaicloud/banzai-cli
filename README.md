:construction: This is a command line interface under heavy development for the [Banzai Cloud Pipeline](https://beta.banzaicloud.io/) platform.

### Installation

```
$ go get github.com/banzaicloud/banzai-cli/cmd/banzai
```

```
$ brew install banzaicloud/tap/banzai-cli
```

### Use

```
A command line client for the Banzai Pipeline platform.

Usage:
  banzai [command]

Available Commands:
  cluster      Manage clusters
  controlplane Manage controlplane
  form         Open forms from config, persist provided values and generate templates
  help         Help about any command
  login        Configure and log in to a Banzai Cloud context
  organization List and select organizations
  secret       Manage secrets

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
      --version              version for banzai

Use "banzai [command] --help" for more information about a command.
```
