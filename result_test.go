package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestResult(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// factory tests
		{scenario: "factory/Accepted",
			exec: func(t *testing.T) {
				// ACT
				result := Accepted()

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: http.StatusAccepted})
			},
		},
		{scenario: "factory/Created",
			exec: func(t *testing.T) {
				// ACT
				result := Created()

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: http.StatusCreated})
			},
		},
		{scenario: "factory/NoContent",
			exec: func(t *testing.T) {
				// ACT
				result := NoContent()

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: http.StatusNoContent})
			},
		},
		{scenario: "factory/OK",
			exec: func(t *testing.T) {
				// ACT
				result := OK()

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: http.StatusOK})
			},
		},
		{scenario: "factory/Status(99)",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidStatusCode).Assert(t)

				// ACT
				_ = Status(99)
			},
		},
		{scenario: "factory/Status(100)",
			exec: func(t *testing.T) {
				// ACT
				result := Status(100)

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: 100})
			},
		},
		{scenario: "factory/Status(599)",
			exec: func(t *testing.T) {
				// ACT
				result := Status(599)

				// ASSERT
				test.That(t, result).Equals(&Result{statusCode: 599})
			},
		},
		{scenario: "factory/Status(600)",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidStatusCode).Assert(t)

				// ACT
				_ = Status(600)
			},
		},

		// String tests
		{scenario: "String/no headers",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{
					statusCode:  404,
					content:     "content",
					contentType: test.AddressOf("content-type"),
				}

				// ACT
				str := sut.String()

				// ASSERT
				test.String(t, str, "result").Equals("{statusCode:404, contentType:content-type, content:[content], headers:<nil>}")
			},
		},
		{scenario: "String/with headers",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{
					"Header-1": "value 1",
					"Header-2": "value 2",
				}}

				// ACT
				str := sut.String()

				// ASSERT
				test.String(t, str, "result").Equals("{statusCode:0, contentType:<nil>, content:[<nil>], headers:{Header-1:[value 1], Header-2:[value 2]}}")
			},
		},

		// method tests
		{scenario: "WithContent",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{}

				// ACT
				sut.WithContent("specified/content-type", []byte("content"))

				// ASSERT
				test.That(t, sut).Equals(&Result{
					content:     []byte("content"),
					contentType: test.AddressOf("specified/content-type"),
				})
			},
		},
		{scenario: "WithHeader/already set",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{
					"Header":   "old value",
					"X-Header": "old x value",
				}}

				// ACT
				_ = sut.WithHeader("Header", "new value")
				_ = sut.WithHeader("x-header", "new x value")

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Header":   "new value",
					"X-Header": "new x value",
				})
			},
		},
		{scenario: "WithHeader/canonical",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithHeader("Canonical-Header", "value")

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Canonical-Header": "value",
				})
			},
		},
		{scenario: "WithHeader/non-canonical",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithHeader("will-be-canonicalised", "value")

				// ASSERT
				test.That(t, sut.headers).Equals(headers{
					"Will-Be-Canonicalised": "value",
				})
			},
		},
		{scenario: "WithHeaders/already set",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{
					"Header-1": "old value 1",
					"Header-2": "old value 2",
				}}

				// ACT
				_ = sut.WithHeaders(map[string]any{
					"Header-2": "new value 2",
					"Header-3": "new value",
				})

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Header-1": "old value 1",
					"Header-2": "new value 2",
					"Header-3": "new value",
				})
			},
		},
		{scenario: "WithHeaders/canonical",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithHeaders(map[string]any{
					"Canonical-String": "value",
					"Canonical-Int":    42,
				})

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Canonical-String": "value",
					"Canonical-Int":    42,
				})
			},
		},
		{scenario: "WithHeaders/non-canonical",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithHeaders(map[string]any{
					"will-be-canonicalised-string": "value",
					"will-be-canonicalised-int":    42,
				})

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Will-Be-Canonicalised-String": "value",
					"Will-Be-Canonicalised-Int":    42,
				})
			},
		},
		{scenario: "WithNonCanonicalHeader",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithNonCanonicalHeader("non-canonical-header", "value")

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"non-canonical-header": "value",
				})
			},
		},
		{scenario: "WithValue",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{}

				// ACT
				sut.WithValue("value")

				// ASSERT
				test.That(t, sut).Equals(&Result{
					content: "value",
				})
			},
		},

		// makeResponse tests
		{scenario: "makeResponse/empty",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{}

				// ACT
				response := sut.makeResponse(&Request{})

				// ASSERT
				test.That(t, response).Equals(&Response{StatusCode: http.StatusOK})
			},
		},
		{scenario: "makeResponse/with content",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{contentType: test.AddressOf("content/type"), content: []byte("content")}

				// ACT
				response := sut.makeResponse(&Request{})

				// ASSERT
				test.That(t, response).Equals(&Response{
					StatusCode:  http.StatusOK,
					ContentType: "content/type",
					Content:     []byte("content"),
				})
			},
		},
		{scenario: "makeResponse/with value",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{
					Accept:         "application/json",
					MarshalContent: json.Marshal,
				}
				sut := &Result{content: "value"}

				// ACT
				response := sut.makeResponse(rq)

				// ASSERT
				test.That(t, response).Equals(&Response{
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Content:     []byte("\"value\""),
				})
			},
		},
		{scenario: "makeResponse/marshal error",
			exec: func(t *testing.T) {
				// ARRANGE
				defer test.Using(&nowUTC, func() time.Time { return time.Date(2010, 9, 8, 7, 6, 5, 0, time.UTC) })()

				var loggedError *InternalError
				defer test.Using(&LogError, func(ei InternalError) { loggedError = &ei })()

				rq := &Request{
					Accept:  "application/json",
					Request: &http.Request{URL: &url.URL{Path: "/path"}},
					MarshalContent: func(v any) ([]byte, error) {
						switch v.(type) {
						case string:
							return nil, errors.New("json.Marshal error")
						default:
							return json.Marshal(v)
						}
					},
				}
				sut := &Result{content: "not used"}

				// ACT
				response := sut.makeResponse(rq)

				// ASSERT
				test.That(t, loggedError).Equals(&InternalError{
					Err:     errors.New("json.Marshal error"),
					Message: "error marshalling Result response",
					Help:    "Result:{statusCode:0, contentType:<nil>, content:[not used], headers:<nil>}",
					Request: rq.Request,
				})

				test.That(t, response.StatusCode).Equals(http.StatusInternalServerError)
				test.That(t, response.ContentType).Equals("application/json")
				test.String(t, response.Content, "content").Equals(`{` +
					`"status":500,` +
					`"error":"Internal Server Error",` +
					`"message":"error marshalling response: json.Marshal error",` +
					`"path":"/path",` +
					`"timestamp":"2010-09-08T07:06:05Z"` +
					`}`)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
