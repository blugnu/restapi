package restapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestHandleRequest(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "successful",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"ID":1,"Name":"test"}`))),
				}

				// ACT
				result := HandleRequest(rq, func(ctx context.Context, c *struct {
					ID   int
					Name string
				}) any {
					test.That(t, c.ID).Equals(1)
					test.That(t, c.Name).Equals("test")
					return http.StatusOK
				})

				// ASSERT
				test.That(t, result).Equals(http.StatusOK)
			},
		},
		{scenario: "request body cannot be read",
			exec: func(t *testing.T) {
				// ARRANGE
				ioerr := errors.New("io error")
				defer test.Using(&ioReadAll, func(io.Reader) ([]byte, error) {
					return nil, ioerr
				})()

				// ACT
				result := HandleRequest(&http.Request{}, func(ctx context.Context, c *struct {
					ID   int
					Name string
				}) any {
					return http.StatusOK
				})

				// ASSERT
				test.Error(t, result.(error)).Is(ErrErrorReadingRequestBody)
			},
		},
		{scenario: "request body is empty",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Body: io.NopCloser(bytes.NewReader([]byte{}))}

				// ACT
				result := HandleRequest(rq, func(ctx context.Context, c *struct {
					ID   int
					Name string
				}) any {
					test.That(t, c).IsNil()
					return http.StatusOK
				})

				// ASSERT
				test.That(t, result).Equals(http.StatusOK)
			},
		},
		{scenario: "request body cannot be unmarshalled",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(`anything`))),
				}
				jsonerr := errors.New("json error")
				defer test.Using(&jsonUnmarshal, func([]byte, interface{}) error {
					return jsonerr
				})()

				// ACT
				result := HandleRequest(rq, func(ctx context.Context, c *struct {
					ID   int
					Name string
				}) any {
					return http.StatusOK
				})

				// ASSERT
				test.Error(t, result.(error)).Is(jsonerr)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
