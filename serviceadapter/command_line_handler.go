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
	potato := commandLineHandler{serviceAdapter: serviceAdapter, logger: logger}
	switch args[1] {
	case "generate-manifest":
		boshInfoJSON := args[2]
		serviceReleasesJSON := args[3]
		planJSON := args[4]
		argsJSON := args[5]
		previousManifestYAML := args[6]
		potato.generateManifest(boshInfoJSON, serviceReleasesJSON, planJSON, argsJSON, previousManifestYAML)
	case "create-binding":
		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		potato.createBinding(bindingID, boshVMsJSON, manifestYAML)
	case "delete-binding":
		bindingID := args[2]
		boshVMsJSON := args[3]
		manifestYAML := args[4]
		potato.deleteBinding(bindingID, boshVMsJSON, manifestYAML)
	default:
		logger.Fatalf("unknown subcommand: %s", args[1])
	}
}

func (p commandLineHandler) generateManifest(boshInfoJSON, serviceReleasesJSON, planJSON, argsJSON, previousManifestYAML string) {
	var boshInfo BoshInfo
	p.must(json.Unmarshal([]byte(boshInfoJSON), &boshInfo), "unmarshalling bosh info")
	p.must(boshInfo.Validate(), "validating bosh info")

	var serviceReleases ServiceReleases
	p.must(json.Unmarshal([]byte(serviceReleasesJSON), &serviceReleases), "unmarshalling service releases")
	p.must(serviceReleases.Validate(), "validating service releases")

	var plan Plan
	p.must(json.Unmarshal([]byte(planJSON), &plan), "unmarshalling service plan")
	p.must(plan.Validate(), "validating service plan")

	var arbitraryParams map[string]interface{}
	p.must(json.Unmarshal([]byte(argsJSON), &arbitraryParams), "unmarshalling arbitraryParams plan")

	var previousManifest *bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(previousManifestYAML), &previousManifest), "unmarshalling previous manifest")

	manifest, err := p.serviceAdapter.GenerateManifest(boshInfo, serviceReleases, plan, arbitraryParams, previousManifest)

	p.mustNot(err, "generating manifest")

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		p.logger.Fatalf("error marshalling bosh manifest: %s", err)
	}

	fmt.Println(string(manifestBytes))
}

func (p commandLineHandler) createBinding(bindingID, boshVMsJSON, manifestYAML string) {
	var boshVMs map[string][]string
	p.must(json.Unmarshal([]byte(boshVMsJSON), &boshVMs), "unmarshalling BOSH VMs")

	var manifest bosh.BoshManifest
	p.must(yaml.Unmarshal([]byte(manifestYAML), &manifest), "unmarshalling manifest")

	binding, err := p.serviceAdapter.CreateBinding(bindingID, boshVMs, manifest)
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
