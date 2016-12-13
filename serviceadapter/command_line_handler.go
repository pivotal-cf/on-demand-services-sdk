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

package serviceadapter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"gopkg.in/yaml.v2"
)

type commandLineHandler struct {
	manifestGenerator     ManifestGenerator
	binder                Binder
	dashboardURLGenerator DashboardUrlGenerator
}

func HandleCommandLineInvocation(args []string, manifestGenerator ManifestGenerator, binder Binder, dashboardUrlGenerator DashboardUrlGenerator) {
	fmt.Fprintf(os.Stderr, "[odb-sdk] handling %s\n", args[1])
	handler := commandLineHandler{manifestGenerator: manifestGenerator, binder: binder, dashboardURLGenerator: dashboardUrlGenerator}
	switch args[1] {
	case "generate-manifest":
		if handler.manifestGenerator != nil {
			serviceDeploymentJSON := args[2]
			planJSON := args[3]
			argsJSON := args[4]
			previousManifestYAML := args[5]
			previousPlanJSON := args[6]
			handler.generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON)
		} else {
			failWithCode(NotImplementedExitCode, "manifest generator not implemented")
		}

	case "create-binding":
		if handler.binder != nil {
			bindingID := args[2]
			boshVMsJSON := args[3]
			manifestYAML := args[4]
			reqParams := args[5]
			handler.createBinding(bindingID, boshVMsJSON, manifestYAML, reqParams)
		} else {
			failWithCode(NotImplementedExitCode, "binder not implemented")
		}
	case "delete-binding":
		if handler.binder != nil {
			bindingID := args[2]
			boshVMsJSON := args[3]
			manifestYAML := args[4]
			unbindingRequestParams := args[5]
			handler.deleteBinding(bindingID, boshVMsJSON, manifestYAML, unbindingRequestParams)
		} else {
			failWithCode(NotImplementedExitCode, "binder not implemented")
		}
	case "dashboard-url":
		if dashboardUrlGenerator != nil {
			instanceID := args[2]
			planJSON := args[3]
			manifestYAML := args[4]
			handler.dashboardUrl(instanceID, planJSON, manifestYAML)
		} else {
			failWithCode(NotImplementedExitCode, "dashboard-url not implemented")
		}
	default:
		fail("unknown subcommand: %s", args[1])
	}
}

func (p commandLineHandler) generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON string) {
	var serviceDeployment ServiceDeployment
	p.must(json.Unmarshal([]byte(serviceDeploymentJSON), &serviceDeployment), "unmarshalling service deployment")
	p.must(serviceDeployment.Validate(), "validating service deployment")

	var plan Plan
	p.must(json.Unmarshal([]byte(planJSON), &plan), "unmarshalling service plan")
	p.must(plan.Validate(), "validating service plan")

	var requestParams map[string]interface{}
	p.must(json.Unmarshal([]byte(argsJSON), &requestParams), "unmarshalling requestParams")

	var previousManifest *bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(previousManifestYAML), &previousManifest), "unmarshalling previous manifest")

	var previousPlan *Plan
	p.must(json.Unmarshal([]byte(previousPlanJSON), &previousPlan), "unmarshalling previous service plan")
	p.must(plan.Validate(), "validating previous service plan")

	manifest, err := p.manifestGenerator.GenerateManifest(serviceDeployment, plan, requestParams, previousManifest, previousPlan)
	if err != nil {
		failWithCodeAndNotifyUser(ErrorExitCode, err.Error())
	}

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		fail("error marshalling bosh manifest: %s", err)
	}

	fmt.Fprintf(os.Stdout, string(manifestBytes))
}

func (p commandLineHandler) createBinding(bindingID, boshVMsJSON, manifestYAML, requestParams string) {
	var boshVMs map[string][]string
	p.must(json.Unmarshal([]byte(boshVMsJSON), &boshVMs), "unmarshalling BOSH VMs")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	var reqParams map[string]interface{}
	p.must(json.Unmarshal([]byte(requestParams), &reqParams), "unmarshalling request binding parameters")

	binding, err := p.binder.CreateBinding(bindingID, boshVMs, manifest, reqParams)
	switch err := err.(type) {
	case BindingAlreadyExistsError:
		failWithCodeAndNotifyUser(BindingAlreadyExistsErrorExitCode, err.Error())
	case AppGuidNotProvidedError:
		failWithCodeAndNotifyUser(AppGuidNotProvidedErrorExitCode, err.Error())
	case error:
		failWithCodeAndNotifyUser(ErrorExitCode, err.Error())
	default:
		break
	}

	p.must(json.NewEncoder(os.Stdout).Encode(binding), "marshalling binding")
}

func (p commandLineHandler) deleteBinding(bindingID, boshVMsJSON, manifestYAML string, requestParams string) {
	var boshVMs bosh.BoshVMs
	p.must(json.Unmarshal([]byte(boshVMsJSON), &boshVMs), "unmarshalling BOSH VMs")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	var params RequestParameters
	p.must(json.Unmarshal([]byte(requestParams), &params), "unmarshalling request binding parameters")

	err := p.binder.DeleteBinding(bindingID, boshVMs, manifest, params)
	switch err.(type) {
	case BindingNotFoundError:
		failWithCodeAndNotifyUser(BindingNotFoundErrorExitCode, err.Error())
	case error:
		failWithCodeAndNotifyUser(ErrorExitCode, err.Error())
	}
}

func (p commandLineHandler) dashboardUrl(instanceID, planJSON, manifestYAML string) {
	var plan Plan
	p.must(json.Unmarshal([]byte(planJSON), &plan), "unmarshalling service plan")
	p.must(plan.Validate(), "validating service plan")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	dashboardUrl, err := p.dashboardURLGenerator.DashboardUrl(instanceID, plan, manifest)
	if err != nil {
		failWithCodeAndNotifyUser(ErrorExitCode, err.Error())
	}

	p.must(json.NewEncoder(os.Stdout).Encode(dashboardUrl), "marshalling dashboardUrl")
}
func (p commandLineHandler) must(err error, msg string) {
	if err != nil {
		fail("error %s: %s\n", msg, err)
	}
}

func (p commandLineHandler) mustNot(err error, msg string) {
	p.must(err, msg)
}

func fail(format string, params ...interface{}) {
	failWithCode(ErrorExitCode, format, params...)
}

func failWithCode(code int, format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("[odb-sdk] %s\n", format), params...)
	os.Exit(code)
}

func failWithCodeAndNotifyUser(code int, format string) {
	fmt.Fprintf(os.Stdout, fmt.Sprintf("%s", format))
	os.Exit(code)
}
