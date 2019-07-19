This is a command line interface under heavy development for the [Banzai Cloud Pipeline](https://beta.banzaicloud.io/) platform.

### Installation

Depending on your environment, you can install it using any of the following methods, using our [DEB](https://banzaicloud.com/downloads/banzai-cli/latest?format=deb) or [RPM](https://banzaicloud.com/downloads/banzai-cli/latest?format=rpm) packages, or binary tarballs for [Linux](https://banzaicloud.com/downloads/banzai-cli/latest?os=linux) and [macOS](https://banzaicloud.com/downloads/banzai-cli/latest?os=darwin).

There is also a shell [script](scripts/getcli.sh) that selects one from the above mentioned methods:

```
$ curl https://getpipeline.sh/cli | sh [-s -- auto|deb|rpm|brew|tar|go|kubectl]
```

You can also select the installation method (one of `auto`, `deb`, `rpm`, `brew`, `tar` or `go`) explicitly, for example:

```
$ curl https://getpipeline.sh/cli | sh -s -- deb
```

You can directly use Homebrew or `go get` as well:

```
$ brew install banzaicloud/tap/banzai-cli
```

```
$ go get github.com/banzaicloud/banzai-cli/cmd/banzai
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
