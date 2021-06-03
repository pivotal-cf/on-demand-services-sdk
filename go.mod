module github.com/pivotal-cf/on-demand-services-sdk

go 1.16

require (
	github.com/google/uuid v1.1.1 // indirect
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pivotal-cf/brokerapi/v8 v8.1.0
	github.com/pkg/errors v0.9.1
	golang.org/x/sys v0.0.0-20210601080250-7ecdf8ef093b // indirect
	golang.org/x/tools v0.1.2 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/yaml.v2 v2.4.0
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
