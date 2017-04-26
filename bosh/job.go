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
	From       string `yaml:"from"`
	Deployment string `yaml:"deployment,omitempty"`
}

func (j Job) AddConsumesLink(name, fromJob string) Job {
	return j.addConsumesLink(name, ConsumesLink{From: fromJob})
}

func (j Job) AddCrossDeploymentConsumesLink(name, fromJob string, deployment string) Job {
	return j.addConsumesLink(name, ConsumesLink{From: fromJob, Deployment: deployment})
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
