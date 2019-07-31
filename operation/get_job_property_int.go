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
// Ex. GetJobPropertyInt("gemfire/tls/enabled")
// gemfire:
//   tls:
//      enabled: 123
//
// Ex. GetJobPropertyInt("gemfire/name=tls/enabled")
// gemfire:
//   - name: tls
//     enabled: 123
func (o *Operation) GetJobPropertyInt(path string) (int, error) {
	if o.error != nil {
		return 0, o.error
	}

	if len(o.jobs) > 1 {
		return 0, errors.New("failed to execute 'GetJobPropertyInt': not implemented for cases where multiple jobs are retrieved")
	}

	var (
		result  int
		entries = strings.Split(path, "/")
	)

	err := o.FetchJobProperty(func(props interface{}) error {
		var err error
		result, err = o.getJobPropertyInt(props, entries, path)
		if err != nil {
			return err
		}

		return nil
	}, entries, path)

	if err != nil {
		return 0, err
	}

	return result, nil
}

func (o *Operation) getJobPropertyInt(props interface{}, entries []string, path string) (int, error) {
	switch ps := props.(type) {
	case map[string]interface{}:
		value := ps[entries[len(entries)-1]]
		v, ok := value.(int)
		if !ok {
			return 0, fmt.Errorf("failed to find int value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	case map[interface{}]interface{}:
		value := ps[entries[len(entries)-1]]
		v, ok := value.(int)
		if !ok {
			return 0, fmt.Errorf("failed to find int value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	case []interface{}:
		value, err := handleIndexingArrayInterface(entries, path, ps)
		if err != nil {
			return 0, err
		}

		v, ok := value.(int)
		if !ok {
			return 0, fmt.Errorf("failed to find int value at '%s', instead '%v'(%T) was found", path, value, value)
		}

		return v, nil
	default:
		// Due to the how fetchJobProperty() works, this code should never be reached
		// presumably...
		return 0, fmt.Errorf("failed to find a supported structure at '%s', instead '%v'(%T) was found", path, props, props)
	}

}
