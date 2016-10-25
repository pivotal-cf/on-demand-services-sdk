package bosh

type BoshManifest struct {
	Name           string                 `yaml:"name"`
	Releases       []Release              `yaml:"releases"`
	Stemcells      []Stemcell             `yaml:"stemcells"`
	InstanceGroups []InstanceGroup        `yaml:"instance_groups"`
	Update         Update                 `yaml:"update"`
	Properties     map[string]interface{} `yaml:"properties,omitempty"`
}

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Stemcell struct {
	Alias   string `yaml:"alias"`
	OS      string `yaml:"os"`
	Version string `yaml:"version"`
}

type InstanceGroup struct {
	Name               string                 `yaml:"name,omitempty"`
	Lifecycle          string                 `yaml:"lifecycle,omitempty"`
	Instances          int                    `yaml:"instances"`
	Jobs               []Job                  `yaml:"jobs,omitempty"`
	VMType             string                 `yaml:"vm_type"`
	VMExtensions       []string               `yaml:"vm_extensions,omitempty"`
	Stemcell           string                 `yaml:"stemcell"`
	PersistentDiskType string                 `yaml:"persistent_disk_type,omitempty"`
	AZs                []string               `yaml:"azs,omitempty"`
	Networks           []Network              `yaml:"networks"`
	Properties         map[string]interface{} `yaml:"properties,omitempty"`
}

type Network struct {
	Name      string   `yaml:"name"`
	StaticIPs []string `yaml:"static_ips,omitempty"`
	Default   []string `yaml:"default,omitempty"`
}

type Update struct {
	Canaries        int    `yaml:"canaries"`
	CanaryWatchTime string `yaml:"canary_watch_time"`
	UpdateWatchTime string `yaml:"update_watch_time"`
	MaxInFlight     int    `yaml:"max_in_flight"`
	Serial          *bool  `yaml:"serial,omitempty"`
}
