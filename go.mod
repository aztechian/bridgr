module github.com/aztechian/bridgr

require (
	github.com/aws/aws-sdk-go v1.43.16
	github.com/briandowns/spinner v1.11.1
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.17+incompatible
	github.com/google/go-cmp v0.5.6
	github.com/mitchellh/mapstructure v1.4.1
	github.com/stretchr/testify v1.7.2
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.9.4
	unknwon.dev/clog/v2 v2.1.2
)

// docker 18.06.1-ce
replace github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20180816081446-320063a2ad06

go 1.13
