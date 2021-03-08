module github.com/pivotal-cf/on-demand-services-sdk

require (
	github.com/google/uuid v1.1.1 // indirect
	github.com/maxbrunsfeld/counterfeiter/v6 v6.3.0
	github.com/onsi/ginkgo v1.15.1
	github.com/onsi/gomega v1.11.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pivotal-cf/brokerapi/v7 v7.5.0
	github.com/pkg/errors v0.9.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/yaml.v2 v2.4.0
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
