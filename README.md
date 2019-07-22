This is a command line interface under heavy development for the [Banzai Cloud Pipeline](https://beta.banzaicloud.io/) platform.

### Installation

Use the following command to quickly install the CLI:

```
$ curl https://getpipeline.sh/cli | sh
```

The [script](scripts/getcli.sh) automatically chooses the best distribution package for your platform.

Available packages:

- [Debian package](https://banzaicloud.com/downloads/banzai-cli/latest?format=deb)
- [RPM package](https://banzaicloud.com/downloads/banzai-cli/latest?format=rpm)
- binary tarballs for [Linux](https://banzaicloud.com/downloads/banzai-cli/latest?os=linux) and [macOS](https://banzaicloud.com/downloads/banzai-cli/latest?os=darwin).

You can also select the installation method (one of `auto`, `deb`, `rpm`, `brew`, `tar` or `go`) explicitly:

```
$ curl https://getpipeline.sh/cli | sh -s -- deb
```

On macOs, you can directly Homebrew:

```
$ brew install banzaicloud/tap/banzai-cli
```

Alternatively, fetch the source and compile it using `go get`:

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
