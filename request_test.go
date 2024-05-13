package restapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/blugnu/test"
)

func TestRequest(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// newRequest tests
		{scenario: "newRequest",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Header: http.Header{}}

				// ACT
				result, err := newRequest(rq)

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, result.Request).Equals(rq)
				test.That(t, result.Accept).Equals("application/json")
				test.That(t, result.MarshalContent).IsNotNil()
			},
		},
		{scenario: "newRequest/unsupported Accept header",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Header: http.Header{"Accept": []string{"text/plain"}}}

				// ACT
				result, err := newRequest(rq)

				// ASSERT
				test.That(t, result).IsNil()
				test.Error(t, err).Is(ErrInvalidAcceptHeader)
			},
		},

		// makeResponse tests
		{scenario: "makeResponse/Error",
			// this test ensures that the request delegates the response creation to the error;
			// detailed tests of error responses in different scenarios are tested in error_test.go
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				err := &Error{}
				isDelegated := false
				defer test.Using(&makeErrorResponse, func(*Error, *Request) *Response {
					isDelegated = true
					return nil
				})()

				// ACT
				_ = rq.makeResponse(err)

				// ASSERT
				test.IsTrue(t, isDelegated)
			},
		},
		{scenario: "makeResponse/Problem",
			// this test ensures that the request delegates the response creation to the problem;
			// detailed tests of problem responses in different scenarios are tested in problem_test.go
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				problem := &Problem{}
				isDelegated := false
				defer test.Using(&makeProblemResponse, func(*Problem, *Request) *Response {
					isDelegated = true
					return nil
				})()

				// ACT
				_ = rq.makeResponse(problem)

				// ASSERT
				test.IsTrue(t, isDelegated)
			},
		},
		{scenario: "makeResponse/Result",
			// this test ensures that the request delegates the response creation to the result;
			// detailed tests of result responses in different scenarios are tested in result_test.go
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				result := &Result{}
				isDelegated := false
				defer test.Using(&makeResultResponse, func(*Result, *Request) *Response {
					isDelegated = true
					return nil
				})()

				// ACT
				_ = rq.makeResponse(result)

				// ASSERT
				test.IsTrue(t, isDelegated)
			},
		},
		{scenario: "makeResponse/error",
			// this test verifies that the request creates an Error of status internal server
			// error response, capturing the details of the error result, and that the response
			// generation is delegated to that Error
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				err := errors.New("test error")
				isDelegated := false
				defer test.Using(&makeErrorResponse, func(e *Error, rq *Request) *Response {
					isDelegated = true
					test.Error(t, e.err).Is(err)
					test.That(t, e.statusCode).Equals(http.StatusInternalServerError)
					return nil
				})()

				// ACT
				_ = rq.makeResponse(err)

				// ASSERT
				test.IsTrue(t, isDelegated)
			},
		},
		{scenario: "makeResponse/[]byte",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				content := []byte("test content")

				// ACT
				response := rq.makeResponse(content)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusOK)
				test.That(t, response.ContentType).Equals("application/octet-stream")
				test.That(t, response.Content).Equals(content)
			},
		},
		{scenario: "makeResponse/[]byte/empty",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				content := []byte{}

				// ACT
				response := rq.makeResponse(content)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNoContent)
			},
		},
		{scenario: "makeResponse/int",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				statusCode := http.StatusAccepted

				// ACT
				response := rq.makeResponse(statusCode)

				// ASSERT
				test.That(t, response.StatusCode).Equals(statusCode)
			},
		},
		{scenario: "makeResponse/default",
			// this test verifies that the request creates a Result encapsulating the result value
			// with a status of OK, delegating the generation of the response to that Result
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{}
				result := "result value"
				isDelegated := false
				defer test.Using(&makeResultResponse, func(r *Result, rq *Request) *Response {
					isDelegated = true
					test.That(t, r.statusCode).Equals(http.StatusOK)
					test.That(t, r.content).Equals(result)
					return nil
				})()

				// ACT
				_ = rq.makeResponse(result)

				// ASSERT
				test.IsTrue(t, isDelegated)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
