package testvariables

import "github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

const (
	OperationFailsKey = "OPERATION_FAILS"

	GenerateManifestServiceDeploymentFileKey = "GENERATE_MANIFEST_SERVICE_DEPLOYMENT_FILE"
	GenerateManifestPlanFileKey              = "GENERATE_MANIFEST_PLAN_FILE"
	GenerateManifestRequestParamsFileKey     = "GENERATE_MANIFEST_REQUEST_PARAMS_FILE"
	GenerateManifestPreviousManifestFileKey  = "GENERATE_MANIFEST_PREVIOUS_MANIFEST_FILE"
	GenerateManifestPreviousPlanFileKey      = "GENERATE_MANIFEST_PREVIOUS_PLAN_FILE"

	BindingIdFileKey       = "BINDING_ID_FILE"
	BindingVmsFileKey      = "BINDING_VMS_FILE"
	BindingManifestFileKey = "BINDING_MANIFEST_FILE"
	BindingParamsFileKey   = "BINDING_PARAMS_FILE"

	DashboardURLInstanceIDKey = "DASHBOARD_URL_INSTANCE_ID_FILE"
	DashboardURLPlanKey       = "DASHBOARD_URL_PLAN_FILE"
	DashboardURLManifestKey   = "DASHBOARD_URL_MANIFEST_FILE"

	DoNotImplementInterfacesKey = "DO_NOT_IMPLEMENT_INTERFACES"

	ErrAppGuidNotProvided   = "no app guid"
	ErrBindingAlreadyExists = "binding already exists"
	ErrBindingNotFound      = "binding not found"
)

var SuccessfulBinding = serviceadapter.Binding{
	RouteServiceURL: "a route",
	SyslogDrainURL:  "a url",
	Credentials: map[string]interface{}{
		"binding": "this binds",
	},
}
