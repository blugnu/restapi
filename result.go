package restapi

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

var (
	makeResultResponse = func(result *Result, rq *Request) *Response {
		response := &Response{
			StatusCode: coalesce(result.statusCode, http.StatusOK),
			headers:    result.headers,
		}

		// if the result content type is non-nil then the corresponding response
		// body is specified in the result content as a []byte
		if result.contentType != nil && result.content != nil {
			response.Content = result.content.([]byte)
			response.ContentType = *result.contentType
			return response
		}

		// a nil content means no further response (no body)
		if result.content == nil {
			return response
		}

		// otherwise the result content holds some value which must be
		// presented in the response according to the request Accept header
		contentType := rq.Accept
		content, err := rq.MarshalContent(result.content)
		if err != nil {
			LogError(InternalError{
				Err:     err,
				Message: "error marshalling Result response",
				Help:    fmt.Sprintf("Result:%v", result),
				Request: rq.Request,
			})
			return InternalServerError(fmt.Errorf("%w: %w", ErrMarshalResultFailed, err)).
				makeResponse(rq)
		}

		response.ContentType = contentType
		response.Content = content
		return response
	}
)

// Result holds details of a valid REST API result.
//
// The Result struct is exported but does not export any members;
// exported methods are provided for endpoint functions to work with
// a Result when required.
//
// An endpoint function will initialise a Result using one of the
// provided functions (e.g. OK, Created, etc.) and then set the
// content and content type and any additional headers if/as required:
//
// # examples
//
//	// return a 200 OK response with a marshalled struct body
//	s := resource{ID: "123", Name: "example"}
//	r := restapi.OK().
//	    WithValue(s)
//
//	// return a 200 OK response with a plain text body
//	// (ignores/overrides any request Accept header)
//	r := restapi.OK().
//	    WithContent("plain/text", []byte("example"))
//
// # methods
type Result struct {
	// content holds the response body content; if nil the response body
	// will be empty
	content any

	// contentType holds the response body content type; if nil the response
	// body content will be marshalled according to the request Accept header
	contentType *string

	// headers holds any additional response headers
	headers headers

	// statusCode holds the HTTP status code for the response
	statusCode int
}

func (h headers) String() string {
	if h == nil {
		return "<nil>"
	}
	items := []string{}
	for k, v := range h {
		items = append(items, fmt.Sprintf("%s:[%v]", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}

// String returns a string representation of the Result.
func (r Result) String() string {
	// if r.headers == nil {
	// 	return fmt.Sprintf("{statusCode:%d, contentType:%s, content:[%v]}",
	// 		r.statusCode, ifNil(r.contentType, "<nil>"), r.content)
	// }

	return fmt.Sprintf("{statusCode:%d, contentType:%v, content:[%v], headers:%v}",
		r.statusCode, ifNil(r.contentType, "<nil>"), r.content, r.headers)
}

// WithContent sets the content and content type of the Result.  The
// specified content and content type will replace any content or
// content type that may have been set on the Result previously.
func (r *Result) WithContent(contentType string, content []byte) *Result {
	r.content = bytes.Clone(content)
	r.contentType = &contentType
	return r
}

// WithHeader sets a canonical header on the Result.
//
// The specified header will be added to any headers already set on the
// Result.  If the specified header is already set on the Result
// the existing header will be replaced with the new value.
//
// The header key is canonicalised using http.CanonicalHeaderKey.  To set
// a header with a non-canonical key use WithNonCanonicalHeader.
func (r *Result) WithHeader(k string, v any) *Result {
	r.headers[http.CanonicalHeaderKey(k)] = v
	return r
}

// WithHeaders sets the headers of the Result.
//
// The specified headers will be added to any headers already set on the
// Result.  If the new headers contain values already set on the Result
// the existing headers will be replaced with the new values.
//
// The header keys are canonicalised using http.CanonicalHeaderKey.
// To set a header with a non-canonical key use WithNonCanonicalHeader.
func (r *Result) WithHeaders(headers map[string]any) *Result {
	for k, v := range headers {
		r.headers[http.CanonicalHeaderKey(k)] = v
	}
	return r
}

// WithNonCanonicalHeader sets a non-canonical header on the Result.
//
// The specified header will be added to any headers already set on the
// Result.  If the specified header is already set on the Result
// the existing header will be replaced with the new value.
//
// The header key is not canonicalised; if the specified key is
// canonical then the canonical header will be set.
//
// WithNonCanonicalHeader should only be used when a non-canonical
// header key is specifically required (which is rare).  Ordinarily
// WithHeader should be used.
func (r *Result) WithNonCanonicalHeader(k string, v any) *Result {
	r.headers[k] = v
	return r
}

// WithValue sets the content of the Result to a value that will be
// marshalled in the response to the content type indicated in the request
// Accept header (or restapi.Default.ResponseContentType).
//
// The specified value will replace any content and content type that may
// have been set on the Result previously.
func (r *Result) WithValue(value any) *Result {
	r.content = value
	r.contentType = nil
	return r
}

// Accepted returns a Result with http.StatusAccepted
func Accepted() *Result { return &Result{statusCode: http.StatusAccepted} }

// Created returns a Result with http.StatusCreated
func Created() *Result { return &Result{statusCode: http.StatusCreated} }

// NoContent returns a Result with http.StatusNoContent
func NoContent() *Result { return &Result{statusCode: http.StatusNoContent} }

// NotImplemented returns a Result with http.StatusNotImplemented
//
// Strictly speaking this is an Error response (in the 5xx range) but is
// provided as a Result as it is a common placeholder response for yet-to-be
// implemented endpoints.  Responses of this nature do not require the
// capabilities of an Error, such as wrapping some runtime error or providing
// additional Help etc.
func NotImplemented() *Result { return &Result{statusCode: http.StatusNotImplemented} }

// OK returns a Result with http.StatusOK
func OK() *Result { return &Result{statusCode: http.StatusOK} }

// Status returns a Result with the specified status code. The status code
// must be in the range 1xx-5xx; any other status code will cause a panic.
//
// # NOTE:
//
// this is a more strict enforcement of standard HTTP response codes than
// is applied by WriteHeader itself which, as of May 2024, accepts codes 1xx-9xx.
func Status(statusCode int) *Result {
	if statusCode < 100 || statusCode > 599 {
		panic(fmt.Errorf("%w: %d: valid range is 1xx-5xx", ErrInvalidStatusCode, statusCode))
	}
	return &Result{statusCode: statusCode}
}

// makeResponse creates a Response from the Result.
//
// If the Result content is nil then the response body will be empty.
//
// If the Result contentType is non-nil then it will be used for the
// Response ContentType and the result content is expected to hold a []byte
// to be used for the Response Content.
//
// If the Result content is non-nil then the response ContentType will
// be set to the request Accept header and the result content will be
// marshalled using the marshalling function provided by the Request.
//
// If the marshalling function returns an error then an InternalServerError
// response will be returned with the error message as the response content;
// an ErrorDetail will also be returned
func (result Result) makeResponse(rq *Request) *Response {
	return makeResultResponse(&result, rq)
}
