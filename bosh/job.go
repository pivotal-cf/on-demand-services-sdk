package bosh

type Job struct {
	Name       string                  `yaml:"name"`
	Release    string                  `yaml:"release"`
	Provides   map[string]ProvidesLink `yaml:"provides,omitempty"`
	Consumes   map[string]interface{}  `yaml:"consumes,omitempty"`
	Properties map[string]interface{}  `yaml:"properties,omitempty"`
}

type ProvidesLink struct {
	As string `yaml:"as"`
}

type ConsumesLink struct {
	From string `yaml:"from"`
}

func (j Job) AddConsumesLink(name, fromJob string) Job {
	return j.addConsumesLink(name, ConsumesLink{From: fromJob})
}

func (j Job) AddNullifiedConsumesLink(name string) Job {
	return j.addConsumesLink(name, "nil")
}

func (j Job) addConsumesLink(name string, value interface{}) Job {
	if j.Consumes == nil {
		j.Consumes = map[string]interface{}{}
	}
	j.Consumes[name] = value
	return j
}
