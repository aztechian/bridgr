module github.com/aztechian/bridgr

require (
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/aws/aws-sdk-go v1.35.23
	github.com/briandowns/spinner v1.16.0
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/google/go-cmp v0.5.3
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.5.1
	rsc.io/letsencrypt v0.0.3 // indirect
	unknwon.dev/clog/v2 v2.1.2
)

// docker 18.06.1-ce
replace github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20180816081446-320063a2ad06

go 1.13
