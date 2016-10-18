package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter/testharness/testvariables"
)

const OperationShouldFail = "true"

func main() {
	for _, envVar := range []string{
		testvariables.GenerateManifestServiceDeploymentFileKey,
		testvariables.GenerateManifestPlanFileKey,
		testvariables.GenerateManifestRequestParamsFileKey,
		testvariables.GenerateManifestPreviousManifestFileKey,
		testvariables.GenerateManifestPreviousPlanFileKey,
		testvariables.BindingIdFileKey,
		testvariables.BindingVmsFileKey,
		testvariables.BindingManifestFileKey,
		testvariables.BindingParamsFileKey,
		testvariables.DashboardURLInstanceIDKey,
		testvariables.DashboardURLPlanKey,
		testvariables.DashboardURLManifestKey,
		testvariables.DoNotImplementInterfacesKey,
	} {
		if os.Getenv(envVar) == "" {
			log.Fatalf("must set %s\n", envVar)
		}
	}

	if os.Getenv(testvariables.DoNotImplementInterfacesKey) == "true" {
		serviceadapter.HandleCommandLineInvocation(os.Args, nil, nil, nil)
		return
	}

	manGen := &manifestGenerator{
		serviceDeploymentFilePath: os.Getenv(testvariables.GenerateManifestServiceDeploymentFileKey),
		planFilePath:              os.Getenv(testvariables.GenerateManifestPlanFileKey),
		requestParamsFilePath:     os.Getenv(testvariables.GenerateManifestRequestParamsFileKey),
		previousManifestFilePath:  os.Getenv(testvariables.GenerateManifestPreviousManifestFileKey),
		previousPlanFilePath:      os.Getenv(testvariables.GenerateManifestPreviousPlanFileKey),
	}

	b := &binder{
		bindingIDFilePath:          os.Getenv(testvariables.BindingIdFileKey),
		deploymentTopologyFilePath: os.Getenv(testvariables.BindingVmsFileKey),
		manifestFilePath:           os.Getenv(testvariables.BindingManifestFileKey),
		requestParamsFilePath:      os.Getenv(testvariables.BindingParamsFileKey),
	}

	d := &dashboard{
		instanceIDFilePath: os.Getenv(testvariables.DashboardURLInstanceIDKey),
		planFilePath:       os.Getenv(testvariables.DashboardURLPlanKey),
		manifestFilePath:   os.Getenv(testvariables.DashboardURLManifestKey),
	}

	serviceadapter.HandleCommandLineInvocation(os.Args, manGen, b, d)
}

type manifestGenerator struct {
	serviceDeploymentFilePath string
	planFilePath              string
	requestParamsFilePath     string
	previousManifestFilePath  string
	previousPlanFilePath      string
}

func (m *manifestGenerator) GenerateManifest(serviceDeployment serviceadapter.ServiceDeployment, plan serviceadapter.Plan, requestParams serviceadapter.RequestParameters, previousManifest *bosh.BoshManifest, previousPlan *serviceadapter.Plan) (bosh.BoshManifest, error) {
	if err := jsonSerialiseToFile(m.serviceDeploymentFilePath, serviceDeployment); err != nil {
		return bosh.BoshManifest{}, err
	}

	if err := jsonSerialiseToFile(m.planFilePath, plan); err != nil {
		return bosh.BoshManifest{}, err
	}

	if err := jsonSerialiseToFile(m.requestParamsFilePath, requestParams); err != nil {
		return bosh.BoshManifest{}, err
	}

	if err := yamlSerialiseToFile(m.previousManifestFilePath, previousManifest); err != nil {
		return bosh.BoshManifest{}, err
	}

	if err := jsonSerialiseToFile(m.previousPlanFilePath, previousPlan); err != nil {
		return bosh.BoshManifest{}, err
	}

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
		}}, nil
}

func jsonSerialiseToFile(filePath string, obj interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(obj)
}

func yamlSerialiseToFile(filePath string, obj interface{}) error {
	objBytes, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, objBytes, 0644)
}

type binder struct {
	bindingIDFilePath          string
	deploymentTopologyFilePath string
	manifestFilePath           string
	requestParamsFilePath      string
}

func (b *binder) CreateBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) (serviceadapter.Binding, error) {
	errs := func(err error) (serviceadapter.Binding, error) {
		return serviceadapter.Binding{}, err
	}

	if err := b.serialiseBindingParams(bindingID, deploymentTopology, manifest, requestParams); err != nil {
		return errs(err)
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
	if err := b.serialiseBindingParams(bindingID, deploymentTopology, manifest, requestParams); err != nil {
		return err
	}

	switch os.Getenv(testvariables.OperationFailsKey) {
	case testvariables.ErrBindingNotFound:
		return serviceadapter.NewBindingNotFoundError(nil)
	case OperationShouldFail:
		return errors.New("An error occurred")
	}

	return nil
}

func (b *binder) serialiseBindingParams(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) error {
	if err := jsonSerialiseToFile(b.bindingIDFilePath, bindingID); err != nil {
		return err
	}

	if err := jsonSerialiseToFile(b.deploymentTopologyFilePath, deploymentTopology); err != nil {
		return err
	}

	if err := yamlSerialiseToFile(b.manifestFilePath, manifest); err != nil {
		return err
	}

	return jsonSerialiseToFile(b.requestParamsFilePath, requestParams)
}

type dashboard struct {
	instanceIDFilePath string
	planFilePath       string
	manifestFilePath   string
}

func (d *dashboard) DashboardUrl(instanceID string, plan serviceadapter.Plan, manifest bosh.BoshManifest) (serviceadapter.DashboardUrl, error) {
	errs := func(err error) (serviceadapter.DashboardUrl, error) {
		return serviceadapter.DashboardUrl{}, err
	}

	if err := jsonSerialiseToFile(d.instanceIDFilePath, instanceID); err != nil {
		return errs(err)
	}

	if err := jsonSerialiseToFile(d.planFilePath, plan); err != nil {
		return errs(err)
	}

	if err := yamlSerialiseToFile(d.manifestFilePath, manifest); err != nil {
		return errs(err)
	}

	if os.Getenv(testvariables.OperationFailsKey) == OperationShouldFail {
		return errs(errors.New("An error occurred"))
	}

	return serviceadapter.DashboardUrl{DashboardUrl: "http://dashboard.com"}, nil
}
