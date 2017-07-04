// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bosh

type BoshManifest struct {
	Name           string                 `yaml:"name"`
	Releases       []Release              `yaml:"releases"`
	Stemcells      []Stemcell             `yaml:"stemcells"`
	InstanceGroups []InstanceGroup        `yaml:"instance_groups"`
	Update         Update                 `yaml:"update"`
	Properties     map[string]interface{} `yaml:"properties,omitempty"`
	Variables      []Variable             `yaml:"variables,omitempty"`
}

type Variable struct {
	Name    string                 `yaml:"name"`
	Type    string                 `yaml:"type"`
	Options map[string]interface{} `yaml:"options,omitempty"`
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
