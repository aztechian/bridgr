module github.com/aztechian/bridgr

require (
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/aws/aws-sdk-go v1.34.9
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/google/go-cmp v0.5.1
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	helm.sh/helm/v3 v3.2.4
)

// docker 18.06.1-ce
replace github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20180816081446-320063a2ad06

go 1.13
