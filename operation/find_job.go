package operation

import "fmt"

// FindJob() is a stateful method that loads a job onto
// the Operations struct. It is useless by itself but
// provides value when chaining update methods on the
// retrieved job
func (o *Operation) FindJob(name string) *Operation {
	if o.error != nil {
		return o
	}

	o.jobs = nil

	for _, ig := range o.manifest.InstanceGroups {
		for i, job := range ig.Jobs {
			if job.Name == name {
				o.jobs = append(o.jobs, &ig.Jobs[i])
			}
		}
	}

	if len(o.jobs) == 0 {
		o.error = fmt.Errorf("failed to find job '%s' within manifest", name)
	}

	return o
}
