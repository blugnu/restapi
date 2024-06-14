package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/blugnu/restapi"
)

// func variables to facilitate testing
var (
	ioReadAll = io.ReadAll
)

// handle reads the request body and unmarshals it into a value of type T
// which is then passed to the supplied function to handle the request.  After
// being read, the request Body is replaced with a new ReadCloser so that it
// may be re-read by the handler function if required.
//
// If the strict argument is true then the request is required to have a non-empty
// body which does not contain any fields not expected by the type T.
//
// If the strict argument is false then an empty body is allowed (the specified
// function will be called with a `nil` argument) and any fields in the request
// body that are not expected by the type T are silently ignored and discarded.
func handle[T any](rq *http.Request, strict bool, h func(c *T) any) any {
	body, err := ioReadAll(rq.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", restapi.ErrErrorReadingRequestBody, err) //NOSONAR
	}
	defer rq.Body.Close()
	rq.Body = io.NopCloser(bytes.NewReader(body))

	if len(body) == 0 {
		// in strict mode, an empty or missing body constitutes a bad request
		if strict {
			return restapi.BadRequest(restapi.ErrBodyRequired)
		}
		// otherwise an empty body may be expected by the handler so we let the
		// handler decide what to do with a 'nil body'
		return h(nil)
	}

	dc := json.NewDecoder(bytes.NewReader(body))
	if strict {
		dc.DisallowUnknownFields()
	}

	c := new(T)
	if err := dc.Decode(c); err != nil {
		// an unpleasant but necessary hack to detect unknown fields since the json
		// decoder does not provide specific errors types which could be used to
		// determine the cause of the error more reliably and precisely :(
		if strings.HasPrefix(err.Error(), "json: unknown field") {
			return restapi.BadRequest(fmt.Errorf("%w: %w", restapi.ErrUnexpectedField, err))
		}
		return fmt.Errorf("%w: %w", ErrDecoder, err)
	}

	return h(c)
}

// HandleRequest reads the request body and unmarshals it into a value of type T
// which is then passed to the supplied function to handle the request.  After
// being read, the request Body is replaced with a new ReadCloser so that it
// may be re-read by the handler function if required.
//
//   - if the request body is empty, the handler function is called with a nil value.
//
//   - if the request body is not empty but cannot be unmarshalled into a value of type T,
//     an error is returned.
//
//   - if the request body contains fields that are not expected by the handler function,
//     they are ignored and discarded
//
// To automatically treat unexpected fields or an empty body as an error, use
// the StrictRequest function.
//
// # example
//
//	func PostResource(rw http.ResponseWriter, rq *http.Request) any {
//	  type resource struct {
//	    Name string `json:"name"`
//	  }
//	  return restapi.HandleRequest(rq, func(r *resource) any {
//	    if r == nil {
//	      return restapi.BadRequest(restapi.ErrBodyRequired)
//	    }
//
//	    // if the request includes an "id" field it is ignored
//	    r.id = uuid.New().String()
//
//	    // ... create a new resource with the required name  ...
//
//	    return restapi.Created().WithValue(r)
//	  })
//	}
func HandleRequest[T any](rq *http.Request, h func(c *T) any) any {
	return handle[T](rq, false, h)
}

// StrictRequest reads the request body and unmarshals it into a value of type T
// which is then passed to the supplied function to handle the request.  After
// being read, the request Body is replaced with a new ReadCloser so that it
// may be re-read by the handler function if required.
//
// The supplied function is not called if:
//
//   - the request body is not empty but cannot be unmarshalled into a value of type T;
//
//   - the request body is empty; restapi.BadRequest(restapi.ErrBodyRequired)
//     is returned;
//
//   - the request body contains fields that are not expected by the marshalled
//     value type; restapi.BadRequest(restapi.ErrUnexpectedField) is returned.
//
// To accept requests with no body or which may contain additional fields not
// supported by the type parameter T, use HandleRequest.
//
// # example
//
//	func PostResource(rw http.ResponseWriter, rq *http.Request) any {
//	   type resource struct {
//	      Name string `json:"name"`
//	   }
//	   return restapi.StrictRequest(rq, func(r *resource) any {
//	      r := struct {
//	         ID string `json:"id"`
//	         resource
//	      }{
//	         ID: uuid.New().String(),
//	         resource: resource {
//	            Name: r.Name,
//	         },
//	      }
//
//	      // ... create the new resource ...
//
//	      return restapi.Created().WithValue(r)
//	   })
//	}
func StrictRequest[T any](rq *http.Request, h func(c *T) any) any {
	return handle[T](rq, true, h)
}
