package restapi

import (
	"fmt"
	"net/http"
)

type Response struct {
	StatusCode  int
	ContentType string
	Content     []byte
	headers
}

// writeResponse writes a response to the http.ResponseWriter.
func (r Response) write(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Add("Content-Type", r.ContentType) //NOSONAR: Content-Type const
	for k, v := range r.headers {
		rw.Header()[k] = []string{fmt.Sprintf("%v", v)}
	}
	rw.WriteHeader(r.StatusCode)
	if err := responseWriterWrite(rw, r.Content); err != nil {
		LogError(InternalError{
			Err:     err,
			Message: "error writing response",
			Help:    fmt.Sprintf("(response: %d %s): rw.Write() error: %s", r.StatusCode, http.StatusText(r.StatusCode), err),
			Request: rq,
		})
	}
}
