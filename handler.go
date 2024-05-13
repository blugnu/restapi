package restapi

import (
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

// EndpointFunc is a function that accepts http.ResponseWriter and
// *http.Request arguments and returns a value of type 'any'.
//
// It is the signature of functions that implement REST API endpoints.
type EndpointFunc func(_ http.ResponseWriter, rq *http.Request) any

// EndpointHandler is an interface that defines a ServeHTTP method that
// accepts http.ResponseWriter and *http.Request arguments and returns a
// value of type 'any'.
//
// This interface must be implemented by types that handle REST API
// endpoint requests.
type EndpointHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) any
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
func HandlerFunc(h EndpointFunc) http.HandlerFunc {
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

		result := h(rw, rq)
		response := makeRequestResponse(apirq, result)
		response.write(rw, rq)
	}
}

// Handler returns a http.HandlerFunc that calls a restapi.EndpointHandler.
//
// A restapi.EndpointHandler is an interface that defines a ServeHTTP method
// that accepts http.ResponseWriter and *http.Request arguments and returns a
// value of type 'any'.
func Handler(h EndpointHandler) http.HandlerFunc {
	return HandlerFunc(h.ServeHTTP)
}
