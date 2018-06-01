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

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/integration_tests/testharness/testvariables"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

const OperationShouldFail = "true"

func main() {
	if os.Getenv(testvariables.DoNotImplementInterfacesKey) == "true" {
		serviceadapter.HandleCLI(os.Args, serviceadapter.CommandLineHandler{})
		return
	}

	handler := serviceadapter.CommandLineHandler{
		ManifestGenerator:     &manifestGenerator{},
		Binder:                &binder{},
		DashboardURLGenerator: &dashboard{},
		SchemaGenerator:       &schemaGenerator{},
	}

	serviceadapter.HandleCLI(os.Args, handler)
}

type manifestGenerator struct{}

func (m *manifestGenerator) GenerateManifest(serviceDeployment serviceadapter.ServiceDeployment, plan serviceadapter.Plan, requestParams serviceadapter.RequestParameters, previousManifest *bosh.BoshManifest, previousPlan *serviceadapter.Plan) (bosh.BoshManifest, error) {
	if os.Getenv(testvariables.OperationFailsKey) == OperationShouldFail {
		fmt.Fprintf(os.Stderr, "not valid")
		return bosh.BoshManifest{}, errors.New("some message to the user")
	}

	return bosh.BoshManifest{Name: "deployment-name",
		Releases: []bosh.Release{
			{
				Name:    "a-release",
				Version: "latest",
			},
		},
		Stemcells: []bosh.Stemcell{
			{
				Alias:   "greatest",
				OS:      "Windows",
				Version: "3.1",
			},
		},
		InstanceGroups: []bosh.InstanceGroup{
			{
				Name: "Test",
				Properties: map[string]interface{}{
					"parseSymbols": "yes%[===",
				},
			},
		},
	}, nil
}

type binder struct{}

func (b *binder) CreateBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters, secrets serviceadapter.ManifestSecrets) (serviceadapter.Binding, error) {
	errs := func(err error) (serviceadapter.Binding, error) {
		return serviceadapter.Binding{}, err
	}

	switch os.Getenv(testvariables.OperationFailsKey) {
	case testvariables.ErrAppGuidNotProvided:
		return errs(serviceadapter.NewAppGuidNotProvidedError(nil))
	case testvariables.ErrBindingAlreadyExists:
		return errs(serviceadapter.NewBindingAlreadyExistsError(nil))
	case OperationShouldFail:
		return errs(errors.New("An internal error occured."))
	}

	return testvariables.SuccessfulBinding, nil
}

func (b *binder) DeleteBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) error {
	switch os.Getenv(testvariables.OperationFailsKey) {
	case testvariables.ErrBindingNotFound:
		return serviceadapter.NewBindingNotFoundError(errors.New("not found"))
	case OperationShouldFail:
		return errors.New("An error occurred")
	}

	return nil
}

type dashboard struct{}

func (d *dashboard) DashboardUrl(instanceID string, plan serviceadapter.Plan, manifest bosh.BoshManifest) (serviceadapter.DashboardUrl, error) {
	if os.Getenv(testvariables.OperationFailsKey) == OperationShouldFail {
		return serviceadapter.DashboardUrl{}, errors.New("An error occurred")
	}

	return serviceadapter.DashboardUrl{DashboardUrl: "http://dashboard.com"}, nil
}

type schemaGenerator struct{}

func (s *schemaGenerator) GeneratePlanSchema(plan serviceadapter.Plan) (serviceadapter.PlanSchema, error) {
	errs := func(err error) (serviceadapter.PlanSchema, error) {
		return serviceadapter.PlanSchema{}, err
	}

	if os.Getenv(testvariables.OperationFailsKey) == OperationShouldFail {
		return errs(errors.New("An error occurred"))
	}

	schemas := serviceadapter.JSONSchemas{
		Parameters: map[string]interface{}{
			"$schema": "http://json-schema.org/draft-04/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"billing-account": map[string]interface{}{
					"description": "Billing account number used to charge use of shared fake server.",
					"type":        "string",
				},
			},
		},
	}
	return serviceadapter.PlanSchema{
		ServiceInstance: serviceadapter.ServiceInstanceSchema{
			Create: schemas,
			Update: schemas,
		},
		ServiceBinding: serviceadapter.ServiceBindingSchema{
			Create: schemas,
		},
	}, nil
}
