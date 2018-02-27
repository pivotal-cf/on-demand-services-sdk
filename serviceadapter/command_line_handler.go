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
	"fmt"
	"io"
	"os"

	"strings"

	"path/filepath"

	"flag"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// CommandLineHandler contains all of the implementers required for the service adapter interface
type CommandLineHandler struct {
	ManifestGenerator     ManifestGenerator
	Binder                Binder
	DashboardURLGenerator DashboardUrlGenerator
	SchemaGenerator       SchemaGenerator
}

type CLIHandlerError struct {
	ExitCode int
	Message  string
}

func (e CLIHandlerError) Error() string {
	return e.Message
}

// Deprecated: Use HandleCLI method of a CommandLineHandler
//
// HandleCommandLineInvocation constructs a CommandLineHandler based on minimal
// service adapter interface handlers and runs HandleCLI based on the
// arguments provided
func HandleCommandLineInvocation(args []string, manifestGenerator ManifestGenerator, binder Binder, dashboardUrlGenerator DashboardUrlGenerator) {
	handler := CommandLineHandler{
		ManifestGenerator:     manifestGenerator,
		Binder:                binder,
		DashboardURLGenerator: dashboardUrlGenerator,
	}
	HandleCLI(args, handler)
}

// HandleCLI calls the correct Service Adapter handler method based on command
// line arguments. The first argument at the command line should be one of:
// generate-manifest, create-binding, delete-binding, dashboard-url.
func HandleCLI(args []string, handler CommandLineHandler) {
	err := handler.Handle(args, os.Stdout, os.Stderr)
	switch e := err.(type) {
	case nil:
	case CLIHandlerError:
		failWithCode(e.ExitCode, err.Error())
	default:
		failWithCode(ErrorExitCode, err.Error())
	}
}

// Handle executes required action and returns an error. Writes responses to the writer provided
func (h CommandLineHandler) Handle(args []string, outputWriter, errorWriter io.Writer) error {
	supportedCommands := h.generateSupportedCommandsMessage()

	if len(args) < 2 {
		return CLIHandlerError{
			ErrorExitCode,
			fmt.Sprintf("the following commands are supported: %s", supportedCommands),
		}
	}

	fmt.Fprintf(errorWriter, "[odb-sdk] handling %s\n", args[1])

	switch args[1] {
	case "generate-manifest":
		if h.ManifestGenerator == nil {
			return CLIHandlerError{NotImplementedExitCode, "manifest generator not implemented"}
		}

		if len(args) < 7 {
			return missingArgsError(args, "<service-deployment-JSON> <plan-JSON> <request-params-JSON> <previous-manifest-YAML> <previous-plan-JSON>")
		}

		serviceDeploymentJSON := args[2]
		planJSON := args[3]
		argsJSON := args[4]
		previousManifestYAML := args[5]
		previousPlanJSON := args[6]
		return h.generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON, outputWriter)

	case "create-binding":
		if h.Binder == nil {
			return CLIHandlerError{NotImplementedExitCode, "binder not implemented"}
		}
		if len(args) < 6 {
			return missingArgsError(args, "<binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>")
		}

		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		reqParams := args[5]
		return h.createBinding(bindingID, boshVMsJSON, manifestYAML, reqParams, outputWriter)
	case "delete-binding":
		if h.Binder == nil {
			return CLIHandlerError{NotImplementedExitCode, "binder not implemented"}
		}
		if len(args) < 6 {
			return missingArgsError(args, "<binding-ID> <bosh-VMs-JSON> <manifest-YAML> <request-params-JSON>")
		}

		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		unbindingRequestParams := args[5]
		return h.deleteBinding(bindingID, boshVMsJSON, manifestYAML, unbindingRequestParams, outputWriter)
	case "dashboard-url":
		if h.DashboardURLGenerator == nil {
			return CLIHandlerError{NotImplementedExitCode, "dashboard-url not implemented"}
		}
		if len(args) < 5 {
			return missingArgsError(args, "<instance-ID> <plan-JSON> <manifest-YAML>")
		}

		instanceID := args[2]
		planJSON := args[3]
		manifestYAML := args[4]
		return h.dashboardUrl(instanceID, planJSON, manifestYAML, outputWriter)
	case "generate-plan-schemas":
		if h.SchemaGenerator == nil {
			return CLIHandlerError{NotImplementedExitCode, "plan schema generator not implemented"}
		}

		planJson, err := parseGeneratePlanSchemaArguments(args, errorWriter)
		if err != nil {
			return err
		}
		return h.generatePlanSchema(planJson, outputWriter)

	default:
		failWithCode(ErrorExitCode, fmt.Sprintf("unknown subcommand: %s. The following commands are supported: %s", args[1], supportedCommands))
	}
	return nil
}

func parseGeneratePlanSchemaArguments(args []string, errorWriter io.Writer) (string, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	flagPlanJSON := fs.String("plan-json", "", "Plan JSON")
	fs.SetOutput(errorWriter)

	err := fs.Parse(args[2:])
	if err != nil {
		return "", incorrectArgsError(args[1])
	}

	if *flagPlanJSON == "" {
		return "", incorrectArgsError(args[1])
	}

	return *flagPlanJSON, nil
}

func failWithMissingArgsError(args []string, argumentNames string) {
	failWithCode(
		ErrorExitCode,
		fmt.Sprintf(
			"Missing arguments for %s. Usage: %s %s %s",
			args[1],
			filepath.Base(args[0]),
			args[1],
			argumentNames,
		),
	)
}

func incorrectArgsError(cmd string) error {
	return CLIHandlerError{
		ErrorExitCode,
		fmt.Sprintf("Incorrect arguments for %s", cmd),
	}
}

func missingArgsError(args []string, argumentNames string) error {
	return CLIHandlerError{
		ExitCode: ErrorExitCode,
		Message: fmt.Sprintf(
			"Missing arguments for %s. Usage: %s %s %s",
			args[1],
			filepath.Base(args[0]),
			args[1],
			argumentNames,
		),
	}
}

func (h CommandLineHandler) generateSupportedCommandsMessage() string {
	var commands []string
	if h.ManifestGenerator != nil {
		commands = append(commands, "generate-manifest")
	}

	if h.Binder != nil {
		commands = append(commands, "create-binding, delete-binding")
	}

	if h.DashboardURLGenerator != nil {
		commands = append(commands, "dashboard-url")
	}

	if h.SchemaGenerator != nil {
		commands = append(commands, "generate-plan-schemas")
	}

	return strings.Join(commands, ", ")
}

func (h CommandLineHandler) generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML, previousPlanJSON string, outputWriter io.Writer) error {
	var serviceDeployment ServiceDeployment

	if err := json.Unmarshal([]byte(serviceDeploymentJSON), &serviceDeployment); err != nil {
		return errors.Wrap(err, "unmarshalling service deployment")
	}
	if err := serviceDeployment.Validate(); err != nil {
		return errors.Wrap(err, "validating service deployment")
	}

	var plan Plan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unmarshalling service plan")
	}
	if err := plan.Validate(); err != nil {
		return errors.Wrap(err, "validating service plan")
	}

	var requestParams map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &requestParams); err != nil {
		return errors.Wrap(err, "unmarshalling requestParams")
	}

	var previousManifest *bosh.BoshManifest
	if err := yaml.Unmarshal([]byte(previousManifestYAML), &previousManifest); err != nil {
		return errors.Wrap(err, "unmarshalling previous manifest")
	}

	var previousPlan *Plan
	if err := json.Unmarshal([]byte(previousPlanJSON), &previousPlan); err != nil {
		return errors.Wrap(err, "unmarshalling previous service plan")
	}
	if previousPlan != nil {
		if err := previousPlan.Validate(); err != nil {
			return errors.Wrap(err, "validating previous service plan")
		}
	}

	manifest, err := h.ManifestGenerator.GenerateManifest(serviceDeployment, plan, requestParams, previousManifest, previousPlan)
	if err != nil {
		fmt.Fprintf(outputWriter, err.Error())
		return CLIHandlerError{ErrorExitCode, err.Error()}
	}

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return errors.Wrap(err, "error marshalling bosh manifest")
	}

	fmt.Fprint(outputWriter, string(manifestBytes))
	return nil
}

func (h CommandLineHandler) createBinding(bindingID, boshVMsJSON, manifestYAML, requestParams string, outputWriter io.Writer) error {
	var boshVMs map[string][]string
	if err := json.Unmarshal([]byte(boshVMsJSON), &boshVMs); err != nil {
		return errors.Wrap(err, "unmarshalling BOSH VMs")
	}

	var manifest bosh.BoshManifest
	if err := yaml.Unmarshal([]byte(manifestYAML), &manifest); err != nil {
		return errors.Wrap(err, "unmarshalling manifest YAML")
	}

	var reqParams map[string]interface{}
	if err := json.Unmarshal([]byte(requestParams), &reqParams); err != nil {
		return errors.Wrap(err, "unmarshalling request binding parameters")
	}

	binding, err := h.Binder.CreateBinding(bindingID, boshVMs, manifest, reqParams)
	if err != nil {
		fmt.Fprintf(outputWriter, err.Error())
		switch err := err.(type) {
		case BindingAlreadyExistsError:
			return CLIHandlerError{BindingAlreadyExistsErrorExitCode, err.Error()}
		case AppGuidNotProvidedError:
			return CLIHandlerError{AppGuidNotProvidedErrorExitCode, err.Error()}
		default:
			return CLIHandlerError{ErrorExitCode, err.Error()}
		}
	}

	if err := json.NewEncoder(outputWriter).Encode(binding); err != nil {
		return errors.Wrap(err, "marshalling binding")
	}

	return nil
}

func (h CommandLineHandler) deleteBinding(bindingID, boshVMsJSON, manifestYAML string, requestParams string, outputWriter io.Writer) error {
	var boshVMs bosh.BoshVMs
	if err := json.Unmarshal([]byte(boshVMsJSON), &boshVMs); err != nil {
		return errors.Wrap(err, "unmarshalling BOSH VMs")
	}

	var manifest bosh.BoshManifest
	if err := yaml.Unmarshal([]byte(manifestYAML), &manifest); err != nil {
		return errors.Wrap(err, "unmarshalling manifest YAML")
	}

	var reqParams RequestParameters
	if err := json.Unmarshal([]byte(requestParams), &reqParams); err != nil {
		return errors.Wrap(err, "unmarshalling request binding parameters")
	}

	err := h.Binder.DeleteBinding(bindingID, boshVMs, manifest, reqParams)
	if err != nil {
		fmt.Fprintf(outputWriter, err.Error())
		switch err.(type) {
		case BindingNotFoundError:
			return CLIHandlerError{BindingNotFoundErrorExitCode, err.Error()}
		default:
			return CLIHandlerError{ErrorExitCode, err.Error()}
		}
	}
	return nil
}

func (h CommandLineHandler) dashboardUrl(instanceID, planJSON, manifestYAML string, outputWriter io.Writer) error {
	var plan Plan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unmarshalling service plan")
	}
	if err := plan.Validate(); err != nil {
		return errors.Wrap(err, "validating service plan")
	}

	var manifest bosh.BoshManifest
	if err := yaml.Unmarshal([]byte(manifestYAML), &manifest); err != nil {
		return errors.Wrap(err, "unmarshalling manifest")
	}

	dashboardUrl, err := h.DashboardURLGenerator.DashboardUrl(instanceID, plan, manifest)
	if err != nil {
		fmt.Fprintf(outputWriter, err.Error())
		return CLIHandlerError{ErrorExitCode, err.Error()}
	}

	if err := json.NewEncoder(outputWriter).Encode(dashboardUrl); err != nil {
		return errors.Wrap(err, "marshalling dashboardUrl")
	}

	return nil
}

func (h CommandLineHandler) generatePlanSchema(planJSON string, outputWriter io.Writer) error {
	var plan Plan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "error unmarshalling plan JSON")
	}
	if err := plan.Validate(); err != nil {
		return errors.Wrap(err, "error validating plan JSON")
	}
	schema, err := h.SchemaGenerator.GeneratePlanSchema(plan)
	if err != nil {
		fmt.Fprintf(outputWriter, err.Error())
		return CLIHandlerError{ErrorExitCode, err.Error()}
	}
	err = json.NewEncoder(outputWriter).Encode(schema)
	if err != nil {
		return errors.Wrap(err, "error marshalling plan schema")
	}

	return nil
}

func (h CommandLineHandler) must(err error, msg string) {
	if err != nil {
		fail("error %s: %s\n", msg, err)
	}
}

func (h CommandLineHandler) mustNot(err error, msg string) {
	h.must(err, msg)
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
