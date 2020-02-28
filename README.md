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

See [command reference](https://banzaicloud.com/docs/pipeline/cli/reference/) in the [official documentation](https://banzaicloud.com/docs/pipeline/cli/).
