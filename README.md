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

### Logging in

To use the command you will have to log in.
You can either log in intaractively using a web browser, or provide an API endpoint and a token manually.

For interactive login, just run `banzai login`, and follow the instructions given.

### Use

A command line client for the Banzai Cloud Pipeline platform.

### Options

```
      --color                use colors on non-tty outputs
      --config string        config file (default is $BANZAICONFIG or $HOME/.banzai/config.yaml)
  -h, --help                 help for banzai
      --interactive          ask questions interactively even if stdin or stdout is non-tty
      --no-color             never display color output
      --no-interactive       never ask questions interactively
      --organization int32   organization id
  -o, --output string        output format (default|yaml|json) (default "default")
      --verbose              more verbose output
```

### SEE ALSO

* [banzai bucket](cmd/docs/banzai_bucket.md)	 - Manage buckets
* [banzai cluster](cmd/docs/banzai_cluster.md)	 - Manage clusters
* [banzai form](cmd/docs/banzai_form.md)	 - Open forms from config, persist provided values and generate templates
* [banzai login](cmd/docs/banzai_login.md)	 - Configure and log in to a Banzai Cloud context
* [banzai organization](cmd/docs/banzai_organization.md)	 - List and select organizations
* [banzai pipeline](cmd/docs/banzai_pipeline.md)	 - Manage deployment of Banzai Cloud Pipeline instances
* [banzai secret](cmd/docs/banzai_secret.md)	 - Manage secrets


For more details, check the [official documentation](http://banzaicloud.com/docs/cli/).
