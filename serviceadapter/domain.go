// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.

// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serviceadapter

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"

	"bytes"

	"gopkg.in/go-playground/validator.v8"
)

//go:generate counterfeiter -o fakes/manifest_generator.go . ManifestGenerator
type ManifestGenerator interface {
	GenerateManifest(serviceDeployment ServiceDeployment, plan Plan, requestParams RequestParameters, previousManifest *bosh.BoshManifest, previousPlan *Plan) (bosh.BoshManifest, error)
}

//go:generate counterfeiter -o fakes/binder.go . Binder
type Binder interface {
	CreateBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams RequestParameters) (Binding, error)
	DeleteBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams RequestParameters) error
}

//go:generate counterfeiter -o fakes/dashboard_url_generator.go . DashboardUrlGenerator
type DashboardUrlGenerator interface {
	DashboardUrl(instanceID string, plan Plan, manifest bosh.BoshManifest) (DashboardUrl, error)
}

//go:generate counterfeiter -o fakes/schema_generator.go . SchemaGenerator
type SchemaGenerator interface {
	GeneratePlanSchema(plan Plan) (PlanSchema, error)
}

type ServiceInstanceSchema struct {
	Create JSONSchemas `json:"create"`
	Update JSONSchemas `json:"update"`
}

type JSONSchemas struct {
	Parameters map[string]interface{} `json:"parameters"`
}

type ServiceBindingSchema struct {
	Create JSONSchemas `json:"create"`
}

type PlanSchema struct {
	ServiceInstance ServiceInstanceSchema `json:"service_instance"`
	ServiceBinding  ServiceBindingSchema  `json:"service_binding"`
}

type DashboardUrl struct {
	DashboardUrl string `json:"dashboard_url"`
}

const (
	ErrorExitCode                     = 1
	NotImplementedExitCode            = 10
	BindingNotFoundErrorExitCode      = 41
	AppGuidNotProvidedErrorExitCode   = 42
	BindingAlreadyExistsErrorExitCode = 49
)

type BindingAlreadyExistsError struct {
	error
}

type AppGuidNotProvidedError struct {
	error
}

type BindingNotFoundError struct {
	error
}

func NewBindingAlreadyExistsError(err error) BindingAlreadyExistsError {
	return BindingAlreadyExistsError{error: fmt.Errorf("binding already exists: %s", err)}
}

func NewAppGuidNotProvidedError(err error) AppGuidNotProvidedError {
	return AppGuidNotProvidedError{error: fmt.Errorf("app GUID not provided: %s", err)}
}

func NewBindingNotFoundError(err error) BindingNotFoundError {
	return BindingNotFoundError{error: fmt.Errorf("binding not found: %s", err)}
}

type RequestParameters map[string]interface{}

func (s RequestParameters) ArbitraryParams() map[string]interface{} {
	if s["parameters"] == nil {
		return map[string]interface{}{}
	}
	return s["parameters"].(map[string]interface{})
}

func (s RequestParameters) ArbitraryContext() map[string]interface{} {
	if s["context"] == nil {
		return map[string]interface{}{}
	}
	return s["context"].(map[string]interface{})
}

func (s RequestParameters) Platform() string {
	context := s.ArbitraryContext()
	platform := context["platform"]
	platformStr, _ := platform.(string)
	return platformStr
}

func (s RequestParameters) BindResource() brokerapi.BindResource {
	marshalledParams, _ := json.Marshal(s["bind_resource"])
	res := brokerapi.BindResource{}
	json.Unmarshal(marshalledParams, &res)
	return res
}

var validate *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}

type ServiceRelease struct {
	Name    string   `json:"name" validate:"required"`
	Version string   `json:"version" validate:"required"`
	Jobs    []string `json:"jobs" validate:"required,min=1"`
}

type ServiceReleases []ServiceRelease

type ServiceDeployment struct {
	DeploymentName string          `json:"deployment_name" validate:"required"`
	Releases       ServiceReleases `json:"releases" validate:"required"`
	Stemcell       Stemcell        `json:"stemcell" validate:"required"`
}

func (r ServiceReleases) Validate() error {
	if len(r) < 1 {
		return errors.New("no releases specified")
	}

	for _, serviceRelease := range r {
		if err := validate.Struct(serviceRelease); err != nil {
			return err
		}
	}

	return nil
}

type Stemcell struct {
	OS      string `json:"stemcell_os" validate:"required"`
	Version string `json:"stemcell_version" validate:"required"`
}

func (s ServiceDeployment) Validate() error {
	return validate.Struct(s)
}

type Properties map[string]interface{}

type Plan struct {
	Properties       Properties       `json:"properties"`
	LifecycleErrands LifecycleErrands `json:"lifecycle_errands,omitempty" yaml:"lifecycle_errands"`
	InstanceGroups   []InstanceGroup  `json:"instance_groups" validate:"required,dive" yaml:"instance_groups"`
	Update           *Update          `json:"update,omitempty"`
}

func (p Plan) Validate() error {
	return validate.Struct(p)
}

type VMExtensions []string

func (e *VMExtensions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ext []string

	if err := unmarshal(&ext); err != nil {
		return err
	}

	for _, s := range ext {
		if s != "" {
			*e = append(*e, s)
		}
	}

	return nil
}

type LifecycleErrands struct {
	PostDeploy []Errand `json:"post_deploy,omitempty" yaml:"post_deploy,omitempty"`
	PreDelete  []Errand `json:"pre_delete,omitempty" yaml:"pre_delete,omitempty"`
}

type Errand struct {
	Name      string   `json:"name,omitempty"`
	Instances []string `json:"instances,omitempty"`
}

type InstanceGroup struct {
	Name               string       `json:"name" validate:"required"`
	VMType             string       `yaml:"vm_type" json:"vm_type" validate:"required"`
	VMExtensions       VMExtensions `yaml:"vm_extensions,omitempty" json:"vm_extensions,omitempty"`
	PersistentDiskType string       `yaml:"persistent_disk_type,omitempty" json:"persistent_disk_type,omitempty"`
	Instances          int          `json:"instances" validate:"min=1"`
	Networks           []string     `json:"networks" validate:"required"`
	AZs                []string     `json:"azs" validate:"required,min=1"`
	Lifecycle          string       `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
	MigratedFrom       []Migration  `yaml:"migrated_from,omitempty" json:"migrated_from,omitempty"`
}

func (i *InstanceGroup) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type aux InstanceGroup

	var tmp aux
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	if tmp.VMExtensions == nil {
		tmp.VMExtensions = VMExtensions{}
	}

	*i = InstanceGroup(tmp)

	return nil
}

type Migration struct {
	Name string `yaml:"name" json:"name"`
}

type Update struct {
	Canaries        int                   `json:"canaries" yaml:"canaries"`
	CanaryWatchTime string                `json:"canary_watch_time" yaml:"canary_watch_time"`
	UpdateWatchTime string                `json:"update_watch_time" yaml:"update_watch_time"`
	MaxInFlight     bosh.MaxInFlightValue `json:"max_in_flight, " yaml:"max_in_flight"`
	Serial          *bool                 `json:"serial,omitempty" yaml:"serial,omitempty"`
}

type updateAlias Update

func (u *Update) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err := decoder.Decode((*updateAlias)(u))
	if err != nil {
		return err
	}

	v, ok := u.MaxInFlight.(json.Number)
	if ok {
		int64v, err := v.Int64()
		if err == nil {
			u.MaxInFlight = int(int64v)
		}
	}

	return bosh.ValidateMaxInFlight(u.MaxInFlight)
}

func (u *Update) MarshalJSON() ([]byte, error) {
	if u.MaxInFlight == nil {
		u.MaxInFlight = 0
	}

	err := bosh.ValidateMaxInFlight(u.MaxInFlight)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal((*updateAlias)(u))
}

func (u *Update) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal((*updateAlias)(u))
	if err != nil {
		return err
	}

	return bosh.ValidateMaxInFlight(u.MaxInFlight)
}

func (u *Update) MarshalYAML() (interface{}, error) {
	err := bosh.ValidateMaxInFlight(u.MaxInFlight)
	if err != nil {
		return []byte{}, err
	}
	return (*updateAlias)(u), nil
}

type Binding struct {
	Credentials     map[string]interface{} `json:"credentials"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
}
