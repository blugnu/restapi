package restapi

// file deepcode ignore XSS: content written to http.ResponseWriter is safe

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	// makeErrorResponse creates a response for an error by projecting the error
	// and marshalling the result according to the acceptable content type for the
	// request.
	//
	// See: func (*Error) makeResponse() for more details.
	//
	// This function is a variable to allow it to be replaced by tests.
	makeErrorResponse = func(e *Error, rq *Request) *Response {
		if e.statusCode != 0 && (e.statusCode < 400 || e.statusCode > 599) {
			panic(fmt.Errorf("%w: %d: valid range for error status is 4xx-5xx", ErrInvalidStatusCode, e.statusCode))
		}

		// apply defaults in case the Error was not fully initialised
		e.statusCode = coalesce(e.statusCode, http.StatusInternalServerError)
		e.timeStamp = coalesce(e.timeStamp, nowUTC())
		e.request = rq.Request

		p := ProjectError(e.info())

		statusCode := e.statusCode
		contentType := rq.Accept
		content, err := rq.MarshalContent(p)
		if err != nil {
			LogError(InternalError{
				Err:     err,
				Message: "error marshalling error response",
				Help:    fmt.Sprintf("the original error was: %v", e),
				Request: rq.Request,
			})
			return &Response{
				StatusCode:  http.StatusInternalServerError,
				ContentType: "plain/text",
				Content: []byte(strings.Join([]string{
					"An error occurred marshalling an error response",
					"",
					"The marshalling error was:",
					"   " + err.Error(),
					"",
					"The original error was:",
					"   " + e.Error(),
				}, "\n")),
			}
		}

		// if returning the intended error response we do NOT include an ErrorInfo
		// result since this signals an error in the error handling process
		return &Response{
			StatusCode:  statusCode,
			ContentType: contentType,
			Content:     content,
		}
	}
)

// Error holds details of a REST API error.
//
// The Error type is exported but has no exported members; an endpoint function
// will usually obtain an Error value using an appropriate factory function, using
// the exported methods to provide information about the error.
//
// # examples
//
//	// an unexpected error occurred
//	if err := SomeOperation(); err != nil {
//	    return restapi.InternalServerError(fmt.Errorf("SomeOperation: %w", err))
//	}
//
//	// an error occurred due to invalid input; provide guidance to the user
//	if err := GetIDFromRequest(rq, &d); err != nil {
//	    return restapi.BadRequest().
//	        WithMessage("ID is missing or invalid").
//	        WithHelp("The ID must be a valid UUID provided in the request path: /v1/resource/<ID>")
//
//	URL         // the URL of the request that resulted in the error
//	TimeStamp   // the (UTC) time that the error occurred
//
// The following additional information may also be provided by a Handler when
// returning an Error:
//
//	Message     // a message to be displayed with the error.  If not provided,
//	            // the message will be the string representation of the error (Err).
//	            //
//	            // NOTE: if Message is set, the Err string will NOT be used
//
//	Help        // a help message to be displayed with the error.  If not provided,
//	            // the help message will be omitted from the response.
type Error struct {
	err        error
	help       *string
	message    *string
	request    *http.Request
	statusCode int
	timeStamp  time.Time
	properties map[string]any
	headers
}

// NewError returns an Error with the specified status code.  One or more additional
// arguments may be provided to be used as follows://
//
//	int        // the status code for the error
//	error      // an error to be wrapped by the Error
//	string     // a message to be displayed with (or instead of) an error
//
// If no status code is provided http.StatusInternalServerError will be used. If
// multiple int arguments are provided only the first will be used; any subsequent
// int arguments will be ignored.
//
// If multiple error arguments are provided they will be wrapped as a single error
// using errors.Join.
//
// If multiple string arguments are provided, the first non-empty string will be used
// as the message; any remaining strings will be ignored.
//
// The returned Error will have a timestamp set to the current time in UTC.
//
// # panics
//
// NewError will panic with the following errors:
//
//   - ErrInvalidArgument if arguments of an unsupported type are provided.
//   - ErrInvalidStatusCode if a status code is provided that is not in the range 4xx-5xx.
//
// # examples
//
//	// no error occurred, but the operation was not successful
//	return NewError(http.StatusNotFound, "no document exists with that ID")
//
//	// an error occurred, but the error is not relevant to the user
//	id, err := GetRequestID(rq)
//	if err != nil {
//	    return NewError(http.BadRequest, "required document ID is missing or invalid", err)
//	}
func NewError(args ...any) *Error {
	var (
		err error
		msg *string
		rq  *http.Request
		sc  *int
	)
	errs := []error{}
	strs := []string{}
	for _, arg := range args {
		switch v := arg.(type) {
		case error:
			errs = append(errs, v)

		case int:
			if v < 400 || v > 599 {
				panic(fmt.Errorf("%w: %d: valid range for error status is 4xx-5xx", ErrInvalidStatusCode, v))
			}
			if sc == nil {
				sc = &v
			}

		case string:
			strs = append(strs, v)

		case *http.Request:
			rq = v
		}
	}

	switch len(errs) {
	case 0:
		// NO-OP
	case 1:
		err = errs[0]
	default:
		err = errors.Join(errs...)
	}

	if len(strs) > 0 {
		s := strings.Join(strs, " ")
		msg = &s
	}

	return &Error{
		statusCode: coalesce(ifNotNil(sc), http.StatusInternalServerError),
		err:        err,
		message:    msg,
		request:    rq,
		timeStamp:  nowUTC(),
	}
}

// info returns an ErrorInfo representing the Error.
func (err *Error) info() ErrorInfo {
	d := ErrorInfo{
		StatusCode: err.statusCode,
		Err:        err.err,
		Message:    ifNotNil(err.message),
		Help:       ifNotNil(err.help),
		Request:    err.request,
		TimeStamp:  err.timeStamp,
	}

	if len(err.properties) > 0 {
		d.Properties = make(map[string]any, len(err.properties))
		for k, v := range err.properties {
			d.Properties[k] = v
		}
	}

	return d
}

// Error implements the error interface for an Error, returning a simplified string
// representation of the error in the form:
//
//	<status code> <status>[: error][: message]
//
// where <status> is the http status text associated with <status code>; <error> and
// <message> are only included if they are set on the Error.
func (err Error) Error() string {
	code := err.statusCode
	status := http.StatusText(err.statusCode)

	switch {
	case err.err != nil && err.message != nil:
		return fmt.Sprintf("%d %s: %s: %s", code, status, err.err, *err.message)

	case err.err != nil:
		return fmt.Sprintf("%d %s: %v", code, status, err.err)

	case err.message != nil:
		return fmt.Sprintf("%d %s: %s", code, status, *err.message)

	default:
		return fmt.Sprintf("%d %s", code, status)
	}
}

// Unwrap returns the error wrapped by the Error (or nil).
func (apierr Error) Unwrap() error {
	return apierr.err
}

// hasHeaders ensures that the Error headers member is an initialised map,
// making a new one if necessary.
func (err *Error) hasHeaders() headers {
	if err.headers == nil {
		err.headers = make(headers)
	}
	return err.headers
}

// hasProperties ensures that the Error properties member is an
// initialised map, making a new one if necessary.
func (err *Error) hasProperties() map[string]any {
	if err.properties == nil {
		err.properties = make(map[string]any)
	}
	return err.properties
}

// makeResponse ensures that the error has a valid status code, timestamp and
// request reference before creating a response by projecting the error
// and marshalling the result according to the acceptable content type for
// the request.
//
// If marshalling fails the response is formed by calling writeError with the
// marshalling error.
//
// An ErrorDetails object is also returned which provides the details of the
// error that was projected.  This object will be passed to the LogError
// function to log the error.
//
// If marshalling the projected error fails, ErrorDetails will describe
// the original error, not the marshalling error.
func (apierr *Error) makeResponse(rq *Request) *Response {
	return makeErrorResponse(apierr, rq)
}

// WithHeader sets a header to be included in the response for the error.
//
// The specified header will be added to any headers already set on the Error.
// If the specified header is already set on the Error the existing header will
// be replaced with the new value.
//
// The header key is canonicalised using http.CanonicalHeaderKey.  To set a header
// with a non-canonical key use WithNonCanonicalHeader.
func (err *Error) WithHeader(k string, v any) *Error {
	err.hasHeaders().set(k, v)
	return err
}

// WithHeaders sets the headers to be included in the response for the error.
//
// The specified headers will be added to any headers already set on the Error.
// If the new headers contain values already set on the Error the existing headers
// will be replaced with the new values.
//
// The header keys are canonicalised using http.CanonicalHeaderKey.  To set a header
// with a non-canonical key use WithNonCanonicalHeader.
func (err *Error) WithHeaders(headers map[string]any) *Error {
	err.hasHeaders().setAll(headers)
	return err
}

// WithHelp sets the help message for the error.
func (err *Error) WithHelp(s string) *Error {
	err.help = &s
	return err
}

// WithMessage sets the message for the error.
func (err *Error) WithMessage(s string) *Error {
	err.message = &s
	return err
}

// WithNonCanonicalHeader sets a non-canonical header to be included in the response
// for the error.
//
// The specified header will be added to any headers already set on the Error.
// If the specified header is already set on the Error the existing header will
// be replaced with the new value.
//
// The header key is not canonicalised; if the specified key is canonical then the
// canonical header will be set.
//
// WithNonCanonicalHeader should only be used when a non-canonical header key is
// specifically required (which is rare).  Ordinarily WithHeader should be used.
func (err *Error) WithNonCanonicalHeader(k string, v any) *Error {
	err.hasHeaders().setNonCanonical(k, v)
	return err
}

// WithProperty sets a property for the error.
func (err *Error) WithProperty(key string, value any) *Error {
	err.hasProperties()[key] = value
	return err
}

// BadRequest returns an ApiError with a status code of 400 and the specified error.
func BadRequest(args ...any) *Error {
	return NewError(append([]any{http.StatusBadRequest}, args...)...)
}

// Forbidden returns an ApiError with a status code of 403 and the specified error.
func Forbidden(args ...any) *Error {
	return NewError(append([]any{http.StatusForbidden}, args...)...)
}

// InternalServerError returns an ApiError with a status code of 500 and the
// specified error.
func InternalServerError(args ...any) *Error {
	return NewError(append([]any{http.StatusInternalServerError}, args...)...)
}

// NotFound returns an ApiError with a status code of 404 and the specified error.
func NotFound(args ...any) *Error {
	return NewError(append([]any{http.StatusNotFound}, args...)...)
}

// Unauthorized returns an ApiError with a status code of 401 and the specified error.
func Unauthorized(args ...any) *Error {
	return NewError(append([]any{http.StatusUnauthorized}, args...)...)
}
