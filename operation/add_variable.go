package operation

import (
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"fmt"
)

// Convenience method to add bosh.Variables to the manifest.
// An error with loaded if the variable already exists.
func (o *Operation) AddVariables(variables ...bosh.Variable) *Operation {
	if o.error != nil {
		return o
	}

	alreadyExists := func(variable bosh.Variable) bool {
		for _, v := range o.manifest.Variables {
			if v.Name == variable.Name {
				return true
			}
		}

		return false
	}

	for _, variable := range variables {
		if alreadyExists(variable) {
			o.error = fmt.Errorf("failed to add bosh.Variable '%s' because it is already present", variable.Name)
			return o
		}

		o.manifest.Variables = append(o.manifest.Variables, variable)
	}

	return o
}