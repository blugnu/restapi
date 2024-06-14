package jsonapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/blugnu/restapi"
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
				result := HandleRequest(rq, func(c *struct {
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
				result := HandleRequest(&http.Request{}, func(c *struct {
					ID   int
					Name string
				}) any {
					return http.StatusOK
				})

				// ASSERT
				test.Error(t, result.(error)).Is(restapi.ErrErrorReadingRequestBody)
			},
		},
		{scenario: "request body is empty",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Body: io.NopCloser(bytes.NewReader([]byte{}))}

				// ACT
				result := HandleRequest(rq, func(c *struct {
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
		{scenario: "invalid json",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(`anything`))),
				}

				// ACT
				result := HandleRequest(rq, func(c *struct {
					ID   int
					Name string
				}) any {
					return http.StatusOK
				})

				// ASSERT
				test.Error(t, result.(error)).Is(ErrDecoder)
			},
		},
		{scenario: "strict request/no body",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Body: io.NopCloser(bytes.NewReader([]byte{}))}

				// ACT
				result := StrictRequest(rq, func(_ *struct {
					Known int
				}) any {
					return http.StatusOK
				})

				// ASSERT
				err, isErr := result.(error)
				test.IsTrue(t, isErr)
				if isErr {
					test.Error(t, err).Is(restapi.BadRequest(restapi.ErrBodyRequired))
				}
			},
		},
		{scenario: "strict request/unknown field in body",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"unknown":1,"Known":2}`))),
				}

				// ACT
				result := StrictRequest(rq, func(_ *struct {
					Known int
				}) any {
					return http.StatusOK
				})

				// ASSERT
				err, isErr := result.(error)
				test.IsTrue(t, isErr)
				if isErr {
					test.Error(t, err).Is(restapi.ErrUnexpectedField)
				}
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
