package restapi

import "net/http"

// InternalError represents an error that occurred during the processing of a request.
//
// Although it is exported this type should not be used directly by REST API
// implementations, except when providing an implementation for the restapi.LogError
// or restapi.ProjectError functions. These functions receive a copy of the Error
// to be logged or projected in the form of an ErrorInfo.
type InternalError struct {
	Err         error
	Help        string
	Message     string
	Request     *http.Request
	ContentType string
}

// LogError is called when an error is returned from a restapi.Handler
// or if an error occurs in an aspect of the restapi implementation itself.
//
// LogError is a function variable with an initial NO-OP implementation,
// i.e. no log is emitted.  Applications should replace the implementation
// with one that produces an appropriate log using the logger configured
// in their application.
var LogError = func(InternalError) { /* NO-OP */ }
