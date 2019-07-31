package operation

import (
	"fmt"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"regexp"
	"strconv"
)

func (o *Operation) FetchJobProperty(handler func(interface{}) error, entries []string, path string, options ...bool) error {
	for _, job := range o.jobs {
		prop, err := o.fetchJobProperty(job, entries, path, options...)
		if err != nil {
			return err
		}

		err = handler(prop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Operation) fetchJobProperty(job *bosh.Job, entries []string, path string, options ...bool) (interface{}, error) {
	autoContructMappings := len(options) > 0 && options[0] == true

	if autoContructMappings && job.Properties == nil {
		job.Properties = map[string]interface{}{}
	}

	var (
		props   interface{} = job.Properties
	)

	for i := 0; i < len(entries)-1; i++ {
		entry := entries[i]

		switch ps := props.(type) {
		case []interface{}:
			v, err := handleArrayInterface(entry, path, ps)
			if err != nil {
				return nil, err
			}

			props = v
		case map[string]interface{}:
			if autoContructMappings && ps[entry] == nil {
				ps[entry] = map[interface{}]interface{}{}
			}

			v := ps[entry]

			if !isSupportedDatastructure(v) {
				return nil, fmt.Errorf(
					"failed to apply property at '%s' because '%v'(%T) exists at .%s",
					path, v, v, entry)
			}

			props = v
		case map[interface{}]interface{}:
			if autoContructMappings && ps[entry] == nil {
				ps[entry] = map[interface{}]interface{}{}
			}

			v := ps[entry]

			if !isSupportedDatastructure(v) {
				return nil, fmt.Errorf(
					"failed to apply property at '%s' because '%v'(%T) exists at .%s",
					path, v, v, entry)
			}

			props = v
		}
	}

	return props, nil
}

func handleArrayInterface(entry string, path string, properties []interface{}) (interface{}, error) {
	var (
		// The only matches we support at this level is querying
		// Ops files can also support indexing at this level
		// Let's defer building that functionality until we need it
		isQueryRegex = regexp.MustCompile(`(.+)=(.+)`)
		matches      = isQueryRegex.FindStringSubmatch(entry)
	)

	if len(matches) != 3 {
		return nil, fmt.Errorf(
			"failed to apply property at '%s' because a query is needed for '%v'(%T) but '%s' was provided",
			path, properties, properties, entry)
	}

	key := matches[1]
	value := matches[2]

	for _, object := range properties {
		switch obj := object.(type) {
		case map[string]interface{}:
			for k, v := range obj {
				if k == key && v == value {
					return object, nil
				}
			}
		case map[interface{}]interface{}:
			for k, v := range obj {
				if k == key && v == value {
					return object, nil
				}
			}
		}
		// there is a minor concern here about catching cases don't match these types
		// but that would be terrible yml for a BOSH manifest.
		// We will handle it if we ever reach that problem
	}

	return nil, fmt.Errorf("failed match '%s' of '%s' in: %v", entry, path, properties)
}

func handleIndexingArrayInterface(entries []string, path string, properties []interface{}) (interface{}, error) {
	var (
		isDigitRegex = regexp.MustCompile(`(\d+)`)
		value        = entries[len(entries)-1]
		matches      = isDigitRegex.FindStringSubmatch(value)
	)

	if len(matches) != 2 {
		return nil, fmt.Errorf("failed to find value at '%s', because '%v'(%T) was found but a non-digit was specified at .%s",
			path, properties, properties, value)
	}

	digit, _ := strconv.Atoi(matches[1])

	if len(properties) <= digit {
		return nil, fmt.Errorf("failed to find value at '%s', because '%v'(%T) only has %d values",
			path, properties, properties, len(properties))
	}

	return properties[digit], nil
}

func isSupportedDatastructure(value interface{}) bool {
	if value == nil {
		return false
	}

	_, ok := value.(map[string]interface{})
	if ok {
		return true
	}

	_, ok = value.(map[interface{}]interface{})
	if ok {
		return true
	}

	_, ok = value.([]interface{})
	return ok
}
