package operation

import (
	"strings"
)

// The method takes a yml path in the form of a unix path and
// any arbitrary value. The null check for o.job is explicitly
// omitted for brevity.
//
// Ex. AddJobProperty("gemfire/tls/enabled", true)
// gemfire:
//   tls:
//      enabled: true
//
// Ex. AddJobProperty("gemfire/name=tls/enabled", true)
// gemfire:
//   - name: tls
//     enabled: true
func (o *Operation) AddJobProperty(path string, value interface{}) *Operation {
	if o.error != nil {
		return o
	}

	entries := strings.Split(path, "/")

	o.error = o.FetchJobProperty(func(props interface{}) error {
		switch ps := props.(type) {
		case map[string]interface{}:
			ps[entries[len(entries)-1]] = value
		case map[interface{}]interface{}:
			ps[entries[len(entries)-1]] = value
			// case []interface{}: is also possible
			// but we haven't had reason to build this
			// functionality (nor an error case) out yet.
		}

		return nil
	}, entries, path, true)

	return o
}
