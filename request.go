package restapi

import (
	"net/http"
)

var (
	// newRequest creates a new Request from an http.Request.  The Accept
	// header is used to determine the content type of the response and the
	// appropriate content marshalling function.
	//
	// If the Accept header is empty or "*/*", the default content type is
	// "application/json".
	//
	// If the Accept header is not recognized or not supported by any
	// content marshalling function, an ErrInvalidAcceptHeader error is
	// returned.
	//
	// newRequest is a function variable to facilitate testing.
	newRequest = func(rq *http.Request) (*Request, error) {
		acc := rq.Header.Get("Accept")
		if acc == "" || acc == "*/*" {
			acc = "application/json"
		}
		if mc, ok := marshal[acc]; ok {
			nrq := &Request{
				Request:        rq,
				Accept:         acc,
				MarshalContent: mc,
			}
			return nrq, nil
		}
		return nil, ErrInvalidAcceptHeader
	}
)

type Request struct {
	*http.Request
	Accept         string
	MarshalContent func(any) ([]byte, error)
}

// makeResponse derives an apppropriate response for a result based on the
// result type as follows:
//
//   - *restapi.Error        // an error response as defined by the Error struct
//   - *restapi.Problem      // an error response as defined by RFC 7807
//   - *restapi.Result       // a successful response as defined by the Result struct
//   - error                 // an internal server error response
//   - []byte                // a byte slice response (Content-Type: application/octet-stream)
//   - int                   // a status code response
//   - <any other type>	     // a successful response with the value marshalled
//     // according to the request Accept header
//
// For []byte values, a 200 OK response is generated with the byte slice as
// the response body with a Content-Type of "application/octet-stream".  If the
// byte slice is empty, the response will have a 204 No Content status code.
//
// For <any other type> responses, the result value is marshalled according
// to the request Accept header (or "application/json" if the Accept header is
// not present, empty or "*/*").  If marshalling of the result value fails,
// a 500 Internal Server Error response is generated.
func (rq *Request) makeResponse(result any) *Response {
	switch result := result.(type) {
	case *Error:
		result.request = rq.Request
		return result.makeResponse(rq)

	case *Result:
		return result.makeResponse(rq)

	case *Problem:
		return result.makeResponse(rq)

	case error:
		return InternalServerError(result, rq.Request).
			makeResponse(rq)

	//FUTURE: option(?) to limit []byte responses to request.Accept = application/octet-stream
	case []byte:
		if len(result) == 0 {
			return &Response{StatusCode: http.StatusNoContent}
		}
		return &Response{
			StatusCode:  http.StatusOK,
			ContentType: "application/octet-stream",
			Content:     result,
		}

	case int:
		return &Response{StatusCode: result}

	default:
		r := Result{
			statusCode: http.StatusOK,
			content:    result,
		}
		return r.makeResponse(rq)
	}
}
