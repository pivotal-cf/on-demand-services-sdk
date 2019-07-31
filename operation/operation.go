package operation

import "github.com/pivotal-cf/on-demand-services-sdk/bosh"

type Operation struct {
	error    error
	manifest *bosh.BoshManifest
	jobs     []*bosh.Job
}

func New(manifest *bosh.BoshManifest) *Operation {
	return &Operation{
		manifest: manifest,
	}
}

func (o *Operation) Error() error {
	return o.error
}
