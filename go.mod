module github.com/aztechian/bridgr

require (
	github.com/aws/aws-sdk-go v1.43.16
	github.com/briandowns/spinner v1.16.0
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.24+incompatible
	github.com/google/go-cmp v0.5.9
	github.com/mitchellh/mapstructure v1.5.0
	github.com/stretchr/testify v1.8.2
	golang.org/x/crypto v0.5.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.11.3
	unknwon.dev/clog/v2 v2.1.2
)

// docker 18.06.1-ce
replace github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20180816081446-320063a2ad06

go 1.13
