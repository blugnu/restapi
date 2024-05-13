package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// func variables to facilitate testing
var (
	ioReadAll     = io.ReadAll
	jsonUnmarshal = json.Unmarshal
)

// HandleRequest reads the request body and unmarshals it into a value of type T
// which is then passed to the supplied function to handle the request.  After
// being read, the request Body is replaced with a new ReadCloser so that it
// may be re-read by the handler function if required.
//
//   - If the request body is empty, the handler function is called with a nil value.
//
//   - If the request body is not empty but cannot be unmarshalled into a value of type T,
//     an error is returned.
//
// # example
//
//	func PostResource(rw http.ResponseWriter, rq *http.Request) any {
//	  type resource struct {
//	    id string   `json:"id"`
//	    name string `json:"name"`
//	  }
//	  return restapi.HandleRequest(rq, func(r *resource) any {
//	    if r == nil {
//	      return restapi.BadRequest(restapi.ErrBodyRequired)
//	    }
//	    r.id = uuid.New().String()
//
//	    // ... create a new resource with the resuiqred name  ...
//
//	    return restapi.Created().WithValue(r)
//	  })
//	}
func HandleRequest[T any](rq *http.Request, h func(ctx context.Context, c *T) any) any {
	ctx := rq.Context()

	body, err := ioReadAll(rq.Body)
	if err != nil {
		return fmt.Errorf("restapi.HandleRequest: %w: %w", ErrErrorReadingRequestBody, err)
	}
	defer rq.Body.Close()
	rq.Body = io.NopCloser(bytes.NewReader(body))

	// an empty body may be expected by the handler so this is not an error;
	// we let the handler decide what to do with a nil body
	if len(body) == 0 {
		return h(ctx, nil)
	}

	c := new(T)
	if err := jsonUnmarshal(body, c); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}

	return h(ctx, c)
}
