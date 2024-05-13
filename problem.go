package restapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	makeProblemResponse = func(p *Problem, rq *Request) *Response {
		response := map[string]any{}
		if p.Type != nil {
			response["type"] = p.Type.String()
		}
		if p.Title != "" {
			response["title"] = p.Title
		}
		if p.Status > 0 {
			response["status"] = p.Status
		}
		if p.Detail != "" {
			response["detail"] = p.Detail
		}
		if p.Instance != nil {
			response["instance"] = p.Instance.String()
		}
		for k, v := range p.props {
			response[k] = v
		}

		if len(response) == 0 {
			panic(fmt.Errorf("%w: an uninitialised restapi.Problem was returned", ErrInvalidOperation))
		}

		result, err := rq.MarshalContent(response)
		if err != nil {
			LogError(InternalError{
				Err:     err,
				Message: "error marshalling Problem response",
				Help:    fmt.Sprintf("Problem: %v", p),
				Request: rq.Request,
			})
			return InternalServerError(err, rq.Request).
				makeResponse(rq)
		}

		return &Response{
			StatusCode:  coalesce(p.Status, http.StatusInternalServerError),
			ContentType: "application/problem+json",
			Content:     result,
		}
	}
)

// FUTURE: implementation of rfc7807; https://datatracker.ietf.org/doc/html/rfc7807

// Implements an RFC7807 Problem Details response
// https://www.rfc-editor.org/rfc/rfc7807
type Problem struct {
	Type     *url.URL
	Status   int
	Instance *url.URL
	Detail   string
	Title    string
	props    map[string]any
}

// NewProblem returns a Problem with the specified arguments. Arguments
// are processed in order and can be of the following types:
//
//	int              // the HTTP status code; will replace any existing Status;
//	                 // if not specified, defaults to http.StatusInternalServerError
//
//	url.URL          // the problem type
//	*url.URL
//
//	string           // the problem detail; will replace any existing detail
//
//	error            // will apply a status code of http.StatusInternalServerError and set the
//	                 // detail to the error message; if the StatusCode or Detail are already
//	                 // set, they will NOT be overwritten
//
//	map[string]any   // additional properties to be included in the response.  If multiple
//	                 // property maps are specified they will be merged; keys from earlier
//	                 // arguments will be overwritten by any values for the same key in later
//	                 // ones
//
// An argument of any other type will cause a panic with ErrInvalidArgument.
//
// If multiple arguments of any of the supported types are specified earlier values in the
// argument list will be applied and over-written by later values (except as noted above).
//
// # examples
//
//	// multiple status codes specified: only the last one is applied
//	NewProblem(http.StatusNotFound, "resource not found", http.BadRequest)
//
// results in a Problem with a StatusCode of 400 (Bad Request) and a Detail of "resource
// not found"
//
//	// status code with multiple errors specified
//	NewProblem(http.StatusBadRequest, errors.New("some error"), errors.New("another error"))
//
// results in a Problem with a StatusCode of 400 (BadRequest) and a Detail of
// "some error" (the second error is ignored)
//
// # note
//
// Some combinations of arguments may result one or more arguments being ignored.  For example,
// specifying a StatusCode, Detail (string) and an error will result in the error being ignored.
func NewProblem(args ...any) *Problem {
	p := Problem{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case int:
			p.Status = arg

		case url.URL:
			p.Type = &arg

		case *url.URL:
			cp := *arg
			p.Type = &cp

		case string:
			p.Detail = arg

		case error:
			p.Status = coalesce(p.Status, http.StatusInternalServerError)
			p.Detail = arg.Error()

		case map[string]any:
			if p.props == nil {
				p.props = map[string]any{}
			}
			for k, v := range arg {
				p.props[k] = v
			}

		default:
			panic(ErrInvalidArgument)
		}
	}
	p.Status = coalesce(p.Status, http.StatusInternalServerError)
	p.Detail = coalesce(p.Detail, http.StatusText(p.Status))
	return &p
}

// makeResponse generates a response for the Problem instance.  The response will be a JSON
// encoded RFC7807 Problem Details response.
//
// If the Problem instance has a Type, Title, Status, Detail or Instance set, these will be
// included in the response.  Any additional properties set on the Problem instance will also
// be included in the response.
func (p *Problem) makeResponse(rq *Request) *Response {
	return makeProblemResponse(p, rq)
}

// WithDetail sets the Detail property of the Problem instance.
//
// The Detail property must provide a human-readable explanation specific to this occurrence
// of the problem.
func (p *Problem) WithDetail(detail string) *Problem {
	p.Detail = detail
	return p
}

// WithInstance sets the instance property of the Problem instance.
//
// The instance property is a URI that identifies the specific occurrence of the problem.
func (p *Problem) WithInstance(instance url.URL) *Problem {
	p.Instance = &instance
	return p
}

// WithProperty sets an additional property on the Problem instance. If the property already
// exists it will be overwritten.
//
// The property name must not be one of the reserved field names:
//
//  - detail
//  - instance
//  - status
//  - title
//  - type
//
// Attempting to specify a property with any of these as key will cause the function to
// panic with ErrInvalidArgument; these fields must be set using the appropriate Problem
// method.

func (p *Problem) WithProperty(key string, value any) *Problem {
	if _, isReserved := map[string]bool{
		"type":     true,
		"title":    true,
		"status":   true,
		"detail":   true,
		"instance": true,
	}[key]; isReserved {
		panic(fmt.Errorf("%w: '%s' is a reserved field which must be set using the appropriate Problem method", ErrInvalidArgument, key))
	}

	if p.props == nil {
		p.props = map[string]any{}
	}
	p.props[key] = value
	return p
}

// WithType sets the Type URL of the Problem instance with an optional Title.  If multiple
// Title values are specified they will be concatenated, with space separators.
func (p *Problem) WithType(url url.URL, title ...string) *Problem {
	p.Type = &url
	p.Title = strings.Join(title, " ")
	return p
}
