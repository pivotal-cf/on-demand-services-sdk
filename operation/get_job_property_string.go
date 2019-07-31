package operation

import (
	"errors"
	"fmt"
	"strings"
)

// The method takes a yml path in the form of a unix path and
// any arbitrary value. The null check for o.job is explicitly
// omitted for brevity.
//
// Ex. GetJobPropertyString("gemfire/tls/enabled")
// gemfire:
//   tls:
//      enabled: "true"
//
// Ex. GetJobPropertyString("gemfire/name=tls/enabled")
// gemfire:
//   - name: tls
//     enabled: true
func (o *Operation) GetJobPropertyString(path string) (string, error) {
	if o.error != nil {
		return "", o.error
	}

	if len(o.jobs) > 1 {
		return "", errors.New("failed to execute 'GetJobPropertyString': not implemented for cases where multiple jobs are retrieved")
	}

	var (
		result  string
		entries = strings.Split(path, "/")
	)

	err := o.FetchJobProperty(func(props interface{}) error {
		var err error
		result, err = o.getJobPropertyString(props, entries, path)
		if err != nil {
			return err
		}

		return nil
	}, entries, path)

	if err != nil {
		return "", err
	}

	return result, nil
}

func (o *Operation) getJobPropertyString(props interface{}, entries []string, path string) (string, error) {
	switch ps := props.(type) {
	case map[string]interface{}:
		value := ps[entries[len(entries)-1]]
		v, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("failed to find string value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	case map[interface{}]interface{}:
		value := ps[entries[len(entries)-1]]
		v, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("failed to find string value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	case []interface{}:
		value, err := handleIndexingArrayInterface(entries, path, ps)
		if err != nil {
			return "", err
		}

		v, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("failed to find string value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	default:
		// Due to the how fetchJobProperty() works, this code should never be reached
		// presumably...
		return "", fmt.Errorf("failed to find a supported structure at '%s', instead '%v'(%T) was found", path, props, props)
	}
}
