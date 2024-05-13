package restapi

// file deepcode ignore XSS: content written to http.ResponseWriter is safe

import (
	"net/http"
	"time"
)

// ErrorInfo represents an error that occurred during the processing of a request.
//
// Although it is exported this type should not be used directly by REST API
// implementations, except when providing an implementation for the restapi.LogError
// or restapi.ProjectError functions. These functions receive a copy of the Error
// to be logged or projected in the form of an ErrorInfo.
type ErrorInfo struct {
	StatusCode int
	Err        error
	Help       string
	Message    string
	Request    *http.Request
	Properties map[string]any
	TimeStamp  time.Time
}
