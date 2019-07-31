package operation

import (
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
)

// Perform an action on the job. Useful for making use
// of provided odb sdk action. The null check for o.job is explicitly
// omitted for brevity.
func (o *Operation) DoJobAction(action func(int, *bosh.Job)) *Operation {
	if o.error != nil {
		return o
	}

	for i, job := range o.jobs {
		action(i, job)
	}

	return o
}
