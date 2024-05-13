package restapi

// file deepcode ignore XSS: content written to http.ResponseWriter is safe

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestError(t *testing.T) {
	// ARRANGE
	// for these tests we fix the time at zero time so that we don't need
	// to initialise timestamp when comparing Error structs
	defer test.Using(&nowUTC, func() time.Time { return time.Time{} })()

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// NewError() tests
		{scenario: "NewError",
			exec: func(t *testing.T) {
				// ARRANGE
				err := errors.New("error")
				rq := &http.Request{}

				// ACT
				result := NewError(404, err, "message", rq)

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					err:        err,
					message:    test.AddressOf("message"),
					request:    rq,
				})
			},
		},
		{scenario: "NewError/0 < status code < 400",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidStatusCode).Assert(t)

				// ACT
				_ = NewError(399, nil)
			},
		},
		{scenario: "NewError/status code > 599",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidStatusCode).Assert(t)

				// ACT
				_ = NewError(600, nil)
			},
		},
		{scenario: "NewError/no args",
			exec: func(t *testing.T) {
				// ACT
				result := NewError()

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 500})
			},
		},
		{scenario: "NewError/multiple errors",
			exec: func(t *testing.T) {
				// ARRANGE
				err1 := errors.New("error 1")
				err2 := errors.New("error 2")

				// ACT
				result := NewError(err1, err2)

				// ASSERT
				test.Error(t, result).Is(err1)
				test.Error(t, result).Is(err2)
				test.That(t, result.err).Equals(errors.Join(err1, err2))
			},
		},
		{scenario: "NewError/multiple strings",
			exec: func(t *testing.T) {
				// ACT
				result := NewError("error 1", "error 2")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 500,
					message:    test.AddressOf("error 1 error 2"),
				})
			},
		},
		{scenario: "NewError/error and message",
			exec: func(t *testing.T) {
				// ARRANGE
				err := errors.New("error")

				// ACT
				result := NewError(err, "message")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 500,
					err:        err,
					message:    test.AddressOf("message"),
				})
			},
		},

		// hasHeaders() / hasProperties() tests
		{scenario: "hasHeaders/nil",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{}

				// ACT
				result := err.hasHeaders()
				result["key"] = "value"

				// ASSERT
				test.That(t, err.headers).IsNotNil()
				test.Map(t, err.headers).Equals(result)
			},
		},
		{scenario: "hasHeaders/not nil",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{headers: headers{"key": "value"}}

				// ACT
				result := err.hasHeaders()

				// ASSERT
				test.Map(t, result).Equals(err.headers)
			},
		},
		{scenario: "hasProperties/nil",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{}

				// ACT
				result := err.hasProperties()
				result["key"] = "value"

				// ASSERT
				test.That(t, err.properties).IsNotNil()
				test.Map(t, err.properties).Equals(result)
			},
		},
		{scenario: "hasProperties/not nil",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{properties: map[string]any{"key": "value"}}

				// ACT
				result := err.hasProperties()

				// ASSERT
				test.Map(t, result).Equals(err.properties)
			},
		},

		// makeResponse() tests
		{scenario: "makeResponse",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404, message: test.AddressOf("message")}
				rq := &Request{
					Request:        &http.Request{URL: &url.URL{}},
					Accept:         "request/Content-Type",
					MarshalContent: func(v any) ([]byte, error) { return []byte("content"), nil },
				}
				var loggedErr *InternalError
				defer test.Using(&LogError, func(e InternalError) { loggedErr = &e })()

				// ACT
				response := err.makeResponse(rq)

				// ASSERT
				test.That(t, loggedErr, "logged error").IsNil()
				test.That(t, response, "response").Equals(&Response{
					StatusCode:  404,
					Content:     []byte("content"),
					ContentType: "request/Content-Type",
				})
			},
		},
		{scenario: "makeResponse/invalid status code",
			exec: func(t *testing.T) {
				// ARRANGE & ASSERT
				err := &Error{statusCode: 600}
				defer test.ExpectPanic(ErrInvalidStatusCode).Assert(t)

				// ACT
				_ = err.makeResponse(nil)
			},
		},
		{scenario: "makeResponse/fails to marshal error response",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404, message: test.AddressOf("message")}
				rq := &Request{
					Request: &http.Request{URL: &url.URL{}},
					Accept:  "application/json",
					MarshalContent: func(v any) ([]byte, error) {
						switch v.(type) {
						case errorResponse:
							return nil, errors.New("marshalling error")
						default:
							return json.Marshal(v)
						}
					},
				}

				var loggedErr *InternalError
				defer test.Using(&LogError, func(e InternalError) { loggedErr = &e })()

				// ACT
				response := err.makeResponse(rq)

				// ASSERT
				test.That(t, loggedErr, "error detail").Equals(&InternalError{
					Err:     errors.New("marshalling error"),
					Message: "error marshalling error response",
					Help:    "the original error was: 404 Not Found: message",
					Request: rq.Request,
				}, "describes the original error")
				test.That(t, response.StatusCode).Equals(500)
				test.That(t, response.ContentType).Equals("plain/text")
				test.Strings(t, response.Content).Equals([]string{
					`An error occurred marshalling an error response`,
					"",
					"The marshalling error was:",
					"   marshalling error",
					"",
					"The original error was:",
					"   404 Not Found: message",
				})
			},
		},

		// Details() tests
		{scenario: "info",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{
					statusCode: 404,
					err:        errors.New("error"),
					help:       test.AddressOf("help"),
					message:    test.AddressOf("message"),
					request:    &http.Request{URL: &url.URL{Path: "/path"}},
					timeStamp:  time.Date(2010, 9, 8, 7, 6, 5, 0, time.UTC),
					properties: map[string]any{"key": "value"},
				}

				// ACT
				result := err.info()

				// ASSERT
				test.That(t, result).Equals(ErrorInfo{
					StatusCode: 404,
					Err:        err.err,
					Message:    "message",
					Help:       "help",
					Request:    err.request,
					Properties: map[string]any{"key": "value"},
					TimeStamp:  time.Date(2010, 9, 8, 7, 6, 5, 0, time.UTC),
				})
			},
		},

		// Error() tests
		{scenario: "Error/no error or message",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.Error()

				// ASSERT
				test.That(t, result).Equals("404 Not Found")
			},
		},
		{scenario: "Error/with error and no message",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404, err: errors.New("error")}

				// ACT
				result := err.Error()

				// ASSERT
				test.That(t, result).Equals("404 Not Found: error")
			},
		},
		{scenario: "Error/with message and no error",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404, message: test.AddressOf("message")}

				// ACT
				result := err.Error()

				// ASSERT
				test.That(t, result).Equals("404 Not Found: message")
			},
		},
		{scenario: "Error/with error and message",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404, err: errors.New("error"), message: test.AddressOf("message")}

				// ACT
				result := err.Error()

				// ASSERT
				test.That(t, result).Equals("404 Not Found: error: message")
			},
		},

		// Unwrap() tests
		{scenario: "Unwrap/no error",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.Unwrap()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "Unwrap/with error",
			exec: func(t *testing.T) {
				// ARRANGE
				e := errors.New("error")
				err := &Error{statusCode: 404, err: e}

				// ACT
				result := err.Unwrap()

				// ASSERT
				test.Error(t, result).Is(e)
			},
		},

		// WithHeader() / WithHeaders() / WithNonCanonicalHeader() tests
		{scenario: "WithHeader",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithHeader("key", "value")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					headers:    headers{"Key": "value"},
				})
			},
		},
		{scenario: "WithHeaders",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithHeaders(map[string]any{
					"key": "value",
				})

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					headers:    headers{"Key": "value"},
				})
			},
		},
		{scenario: "WithNonCanonicalHeader",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithNonCanonicalHeader("key", "value")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					headers:    headers{"key": "value"},
				})
			},
		},

		// WithHelp() / WithMessage() tests
		{scenario: "WithHelp",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithHelp("help")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					help:       test.AddressOf("help"),
				})
			},
		},
		{scenario: "WithMessage",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithMessage("message")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					message:    test.AddressOf("message"),
				})
			},
		},

		// WithProperty() test
		{scenario: "WithProperty",
			exec: func(t *testing.T) {
				// ARRANGE
				err := &Error{statusCode: 404}

				// ACT
				result := err.WithProperty("key", "value")

				// ASSERT
				test.That(t, result).Equals(&Error{
					statusCode: 404,
					properties: map[string]any{"key": "value"},
				})
			},
		},

		// factory tests
		{scenario: "factory/Error",
			exec: func(t *testing.T) {
				// ARRANGE
				err := errors.New("error")

				// this test ensures that the Error() factory applies the current time
				// to the timestamp field
				ts := time.Now()
				defer test.Using(&nowUTC, func() time.Time { return ts })()

				// ACT
				result := NewError(404, err)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 404, err: err, timeStamp: ts})
			},
		},
		{scenario: "factory/BadRequest",
			exec: func(t *testing.T) {
				// ACT
				result := BadRequest(nil)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 400})
			},
		},
		{scenario: "factory/Forbidden",
			exec: func(t *testing.T) {
				// ACT
				result := Forbidden(nil)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 403})
			},
		},
		{scenario: "factory/InternalServerError",
			exec: func(t *testing.T) {
				// ACT
				result := InternalServerError(nil)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 500})
			},
		},
		{scenario: "factory/NotFound",
			exec: func(t *testing.T) {
				// ACT
				result := NotFound(nil)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 404})
			},
		},
		{scenario: "factory/Unauthorized",
			exec: func(t *testing.T) {
				// ACT
				result := Unauthorized(nil)

				// ASSERT
				test.That(t, result).Equals(&Error{statusCode: 401})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
