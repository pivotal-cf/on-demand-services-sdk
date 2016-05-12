package serviceadapter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"gopkg.in/yaml.v2"
)

type commandLineHandler struct {
	serviceAdapter ServiceAdapter
	logger         *log.Logger
}

func HandleCommandLineInvocation(args []string, serviceAdapter ServiceAdapter, logger *log.Logger) {
	logger.Printf("handling %s", args[1])
	handler := commandLineHandler{serviceAdapter: serviceAdapter, logger: logger}
	switch args[1] {
	case "generate-manifest":
		serviceDeploymentJSON := args[2]
		planJSON := args[3]
		argsJSON := args[4]
		previousManifestYAML := args[5]
		handler.generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML)
	case "create-binding":
		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		bindingArbitraryParams := args[5]
		handler.createBinding(bindingID, boshVMsJSON, manifestYAML, bindingArbitraryParams)
	case "delete-binding":
		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		handler.deleteBinding(bindingID, boshVMsJSON, manifestYAML)
	default:
		logger.Fatalf("unknown subcommand: %s", args[1])
	}
}

func (p commandLineHandler) generateManifest(serviceDeploymentJSON, planJSON, argsJSON, previousManifestYAML string) {

	var serviceDeployment ServiceDeployment
	p.must(json.Unmarshal([]byte(serviceDeploymentJSON), &serviceDeployment), "unmarshalling service deployment")
	p.must(serviceDeployment.Validate(), "validating service deployment")

	var plan Plan
	p.must(json.Unmarshal([]byte(planJSON), &plan), "unmarshalling service plan")
	p.must(plan.Validate(), "validating service plan")

	var arbitraryParams map[string]interface{}
	p.must(json.Unmarshal([]byte(argsJSON), &arbitraryParams), "unmarshalling arbitraryParams plan")

	var previousManifest *bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(previousManifestYAML), &previousManifest), "unmarshalling previous manifest")

	manifest, err := p.serviceAdapter.GenerateManifest(serviceDeployment, plan, arbitraryParams, previousManifest)

	p.mustNot(err, "generating manifest")

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		p.logger.Fatalf("error marshalling bosh manifest: %s", err)
	}

	fmt.Println(string(manifestBytes))
}

func (p commandLineHandler) createBinding(bindingID, boshVMsJSON, manifestYAML, arbitraryParams string) {
	var boshVMs map[string][]string
	p.must(json.Unmarshal([]byte(boshVMsJSON), &boshVMs), "unmarshalling BOSH VMs")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	var params map[string]interface{}
	p.must(json.Unmarshal([]byte(arbitraryParams), &params), "unmarshalling arbitrary binding parameters")

	binding, err := p.serviceAdapter.CreateBinding(bindingID, boshVMs, manifest, params)
	p.mustNot(err, "creating binding")

	p.must(json.NewEncoder(os.Stdout).Encode(binding), "marshalling binding")
}

func (p commandLineHandler) deleteBinding(bindingID, boshVMsJSON, manifestYAML string) {
	var boshVMs bosh.BoshVMs
	p.must(json.Unmarshal([]byte(boshVMsJSON), &boshVMs), "unmarshalling BOSH VMs")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	err := p.serviceAdapter.DeleteBinding(bindingID, boshVMs, manifest)
	p.mustNot(err, "deleting binding")
}

func (p commandLineHandler) must(err error, msg string) {
	if err != nil {
		p.logger.Fatalf("error %s: %s\n", msg, err)
	}
}

func (p commandLineHandler) mustNot(err error, msg string) {
	p.must(err, msg)
}
