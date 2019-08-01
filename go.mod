module github.com/banzaicloud/banzai-cli

require (
	git.apache.org/thrift.git v0.0.0-20180902110319-2566ecd5d999 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/antihax/optional v0.0.0-20180407024304-ca021399b1a6
	github.com/banzaicloud/backyards-cli v0.0.0-20190807085854-e629b193f784
	github.com/coreos/go-oidc v2.0.0+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/buffalo-plugins v1.13.0 // indirect
	github.com/gobuffalo/meta v0.0.0-20190207205153-50a99e08b8cf // indirect
	github.com/gobuffalo/packr/v2 v2.0.2
	github.com/google/uuid v1.1.1
	github.com/goph/emperror v0.17.2
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.4
	github.com/microcosm-cc/bluemonday v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/skratchdot/open-golang v0.0.0-20190104022628-a2dfa6d0dab6
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/ttacon/chalk v0.0.0-20140724125006-76b3c8b611de
	golang.org/x/build v0.0.0-20190111050920-041ab4dc3f9d // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/AlecAivazis/survey.v1 v1.8.2
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	sigs.k8s.io/kind v0.4.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
