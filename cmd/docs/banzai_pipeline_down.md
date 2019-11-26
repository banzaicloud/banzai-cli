## banzai pipeline down

Destroy the controlplane

### Synopsis

Destroy a controlplane based on json stdin or interactive session

```
banzai pipeline down [flags]
```

### Options

```
      --auto-approve               Automatically approve the changes to deploy (default true)
      --container-runtime string   Run the terraform command with "docker", "containerd" (crictl) or "exec" (execute locally) (default "auto")
  -h, --help                       help for down
      --image string               Name of Docker image repository to use (default "docker.io/banzaicloud/pipeline-installer")
      --image-pull                 Pull installer image even if it's present locally (default true)
      --image-tag string           Tag of installer Docker image to use (default "latest")
      --workspace string           Name of directory for storing the applied configuration and deployment status
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

* [banzai pipeline](banzai_pipeline.md)	 - Manage deployment of Banzai Cloud Pipeline instances

