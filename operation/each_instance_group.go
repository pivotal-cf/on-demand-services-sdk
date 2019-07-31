package operation

import "github.com/pivotal-cf/on-demand-services-sdk/bosh"

// Perform an action on each instance group. Useful for making use
// of provided odb sdk action.
func (o *Operation) EachInstanceGroup(action func(*bosh.InstanceGroup)) *Operation {
	if o.error != nil {
		return o
	}

	for i, _ := range o.manifest.InstanceGroups {
		action(&o.manifest.InstanceGroups[i])
	}

	return o
}

