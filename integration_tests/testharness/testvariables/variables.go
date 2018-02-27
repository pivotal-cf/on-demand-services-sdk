package testvariables

import "github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"

const (
	OperationFailsKey = "OPERATION_FAILS"

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
