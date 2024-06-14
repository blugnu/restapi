package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// function variables to facilitate testing
var (
	responseWriterWrite = func(rw http.ResponseWriter, content []byte) error {
		_, err := rw.Write(content)
		return err
	}

	makeRequestResponse = func(r *Request, result any) *Response {
		return r.makeResponse(result)
	}
)

// EndpointFunc is a type for a function that conforms to the EndpointHandler
// ServeAPI() method.  A function with the appropriate signature may be
// converted to an EndpointHandler by casting to this type.
//
// # example
//
//	func MyEndpoint(ctx context.Context, rq *http.Request) any {
//		// do something
//		return result
//	}
//
//	var MyHandler = EndpointFunc(MyEndpoint)
//
//	func main() {
//		http.HandleFunc("/my-endpoint", restapi.Handler(MyHandler))
//		http.ListenAndServe(":8080", nil)
//	}
type EndpointFunc func(context.Context, *http.Request) any

// ServeAPI implements the EndpointHandler interface for the EndpointFunc type.
func (f EndpointFunc) ServeAPI(ctx context.Context, rq *http.Request) any {
	return f(ctx, rq)
}

// EndpointHandler is an interface that defines a ServeAPI method that
// conforms to the EndpointFunc signature, accepting a context.Context and
// *http.Request arguments, returning a value of type 'any'.
type EndpointHandler interface {
	ServeAPI(context.Context, *http.Request) any
}

// HandlerFunc returns a http.HandlerFunc that calls a REST API endpoint
// function.
//
// The endpoint function differs from a http handler function in that
// in addition to accepting http.ResponseWriter and *http.Request
// arguments, it also returns a value of type 'any'.
//
// The returned value is processed by the Handler function to generate
// an appropriate response.
func HandlerFunc(h func(context.Context, *http.Request) any) http.HandlerFunc {
	return func(rw http.ResponseWriter, rq *http.Request) {
		apirq, err := newRequest(rq)
		if err != nil {
			content, _ := json.Marshal(err.Error())
			statusCode := http.StatusInternalServerError
			if errors.Is(err, ErrInvalidAcceptHeader) {
				//FUTURE: produce the list of supported content types from the marshal map
				statusCode = http.StatusNotAcceptable
				content = []byte("[\"application/json\",\"application/xml\",\"text/json\",\"test/xml\",\"*/*\",none]")
			}
			LogError(InternalError{
				Err:     err,
				Request: rq,
				Message: "error initialising request",
			})
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(statusCode)
			if rwerr := responseWriterWrite(rw, content); rwerr != nil {
				LogError(InternalError{
					Err:     rwerr,
					Message: "error writing request error response",
					Help:    fmt.Sprintf("(request error: %s): rw.Write() error: %s", err, rwerr),
					Request: rq,
				})
			}
			return
		}

		defer func() {
			if r := recover(); r != nil {
				LogError(InternalError{
					Err:     fmt.Errorf("%v", r),
					Message: "handler panic",
					Request: rq,
				})
				InternalServerError(fmt.Errorf("panic: %v", r)).
					makeResponse(apirq).
					write(rw, rq)
			}
		}()

		result := h(rq.Context(), rq)
		response := makeRequestResponse(apirq, result)
		response.write(rw, rq)
	}
}

// Handler returns a http.HandlerFunc that calls a restapi.EndpointHandler.
//
// A restapi.EndpointHandler is an interface that defines a ServeAPI method
// that accepts a context.Context and a *http.Request argument, returning a
// value of type 'any'.
func Handler(h EndpointHandler) http.HandlerFunc {
	return HandlerFunc(h.ServeAPI)
}
