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

func TestProblem(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "newProblem/no args",
			exec: func(t *testing.T) {
				// ACT
				result := NewProblem()

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "Internal Server Error",
				})
			},
		},
		{scenario: "newProblem/http status code",
			exec: func(t *testing.T) {
				// ACT
				result := NewProblem(http.StatusBadRequest)

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusBadRequest,
					Detail: "Bad Request",
				})
			},
		},
		{scenario: "newProblem/http status code and title",
			exec: func(t *testing.T) {
				// ACT
				result := NewProblem(http.StatusNotFound, "the resource could not be found")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusNotFound,
					Detail: "the resource could not be found",
				})
			},
		},
		{scenario: "newProblem/error",
			exec: func(t *testing.T) {
				// ACT
				err := errors.New("some error")
				result := NewProblem(err)

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "some error",
				})
			},
		},
		{scenario: "newProblem/error and status code",
			exec: func(t *testing.T) {
				// ACT
				err := errors.New("some error")
				result := NewProblem(err, http.StatusForbidden)

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusForbidden,
					Detail: "some error",
				})
			},
		},
		{scenario: "newProblem/error and status code and detail",
			exec: func(t *testing.T) {
				// ACT
				err := errors.New("some error")
				result := NewProblem(err, http.StatusForbidden, "forbidden")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusForbidden,
					Detail: "forbidden",
				})
			},
		},
		{scenario: "newProblem/url (by value)",
			exec: func(t *testing.T) {
				// ACT
				urlval := url.URL{Scheme: "http", Host: "example.com"}
				result := NewProblem(urlval)

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "Internal Server Error",
					Type:   &url.URL{Scheme: "http", Host: "example.com"},
				})

				urlval.Scheme = "https"
				test.That(t, result.Type.Scheme, "Problem.Type is a copy of the URL").Equals("http")
			},
		},
		{scenario: "newProblem/url (by ref)",
			exec: func(t *testing.T) {
				// ACT
				urlref := &url.URL{Scheme: "http", Host: "example.com"}
				result := NewProblem(urlref)

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "Internal Server Error",
					Type:   &url.URL{Scheme: "http", Host: "example.com"},
				})

				urlref.Scheme = "https"
				test.That(t, result.Type.Scheme, "Problem.Type is a copy of the URL").Equals("http")
			},
		},
		{scenario: "newProblem/props",
			exec: func(t *testing.T) {
				// ACT
				result := NewProblem(map[string]any{
					"prop1": "value1",
					"prop2": 2,
				})

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "Internal Server Error",
					props: map[string]any{
						"prop1": "value1",
						"prop2": 2,
					},
				})
			},
		},
		{scenario: "newProblem/multiple props",
			exec: func(t *testing.T) {
				// ACT
				result := NewProblem(
					map[string]any{
						"prop1": "value1",
						"prop2": 2,
					},
					map[string]any{
						"prop2": "value2",
						"prop3": 3,
					})

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Status: http.StatusInternalServerError,
					Detail: "Internal Server Error",
					props: map[string]any{
						"prop1": "value1",
						"prop2": "value2",
						"prop3": 3,
					},
				})
			},
		},
		{scenario: "newProblem/invalid argument",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidArgument).Assert(t)

				// ACT
				_ = NewProblem(false)
			},
		},

		// With...() tests
		{scenario: "WithDetail",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithDetail("some detail")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Detail: "some detail",
				})
			},
		},
		{scenario: "WithInstance",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithInstance(url.URL{Scheme: "http", Host: "example.com"})

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Instance: &url.URL{Scheme: "http", Host: "example.com"},
				})
			},
		},
		{scenario: "WithProperty",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithProperty("prop1", "value1")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					props: map[string]any{"prop1": "value1"},
				})
			},
		},
		{scenario: "WithProperty/using reserved field name",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrInvalidArgument).Assert(t)

				// ACT
				_ = (&Problem{}).WithProperty("type", "some value")
			},
		},
		{scenario: "WithType/no title",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithType(url.URL{Scheme: "http", Host: "example.com"})

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Type: &url.URL{Scheme: "http", Host: "example.com"},
				})
			},
		},
		{scenario: "WithType/with single title arg",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithType(url.URL{Scheme: "http", Host: "example.com"}, "some title")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Type:  &url.URL{Scheme: "http", Host: "example.com"},
					Title: "some title",
				})
			},
		},
		{scenario: "WithType/with multiple title args",
			exec: func(t *testing.T) {
				// ARRANGE
				p := &Problem{}

				// ACT
				result := p.WithType(url.URL{Scheme: "http", Host: "example.com"}, "some", "title")

				// ASSERT
				test.That(t, result).Equals(&Problem{
					Type:  &url.URL{Scheme: "http", Host: "example.com"},
					Title: "some title",
				})
			},
		},

		// makeResponse() tests
		{scenario: "makeResponse/zero value",
			exec: func(t *testing.T) {
				// ARRANGE + ASSERT
				rq := &Request{}
				defer test.ExpectPanic(ErrInvalidOperation).Assert(t)

				// ACT
				_ = (&Problem{}).makeResponse(rq)
			},
		},
		{scenario: "makeResponse/with status code",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{MarshalContent: json.Marshal}

				// ACT
				response := (&Problem{Status: http.StatusNotFound}).makeResponse(rq)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNotFound)
				test.String(t, response.Content).Equals("{\"status\":404}")
			},
		},
		{scenario: "makeResponse/with status code and detail",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{MarshalContent: json.Marshal}

				// ACT
				response := (&Problem{Status: http.StatusNotFound, Detail: "not found"}).makeResponse(rq)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNotFound)
				test.String(t, response.Content).Equals("{\"detail\":\"not found\",\"status\":404}")
			},
		},
		{scenario: "makeResponse/with status code, type and title",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{MarshalContent: json.Marshal}

				// ACT
				response := (&Problem{
					Status: http.StatusNotFound,
					Type:   &url.URL{Scheme: "http", Host: "example.com"},
					Title:  "the requested resource was not found",
				}).makeResponse(rq)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNotFound)
				test.String(t, response.Content).Equals("{\"status\":404,\"title\":\"the requested resource was not found\",\"type\":\"http://example.com\"}")
			},
		},
		{scenario: "makeResponse/with status code and instance",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{MarshalContent: json.Marshal}

				// ACT
				response := (&Problem{
					Status:   http.StatusNotFound,
					Instance: &url.URL{Scheme: "http", Host: "example.com"},
				}).makeResponse(rq)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNotFound)
				test.String(t, response.Content).Equals("{\"instance\":\"http://example.com\",\"status\":404}")
			},
		},
		{scenario: "makeResponse/with status code and props",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &Request{MarshalContent: json.Marshal}

				// ACT
				response := (&Problem{
					Status: http.StatusNotFound,
					props: map[string]any{
						"prop1": "value1",
						"prop2": 2,
					},
				}).makeResponse(rq)

				// ASSERT
				test.That(t, response.StatusCode).Equals(http.StatusNotFound)
				test.String(t, response.Content).Equals("{\"prop1\":\"value1\",\"prop2\":2,\"status\":404}")
			},
		},
		{scenario: "makeResponse/marshalling error",
			exec: func(t *testing.T) {
				// ARRANGE
				errm := errors.New("marshalling error")
				rq := &Request{
					Accept:  "request/Content-Type",
					Request: &http.Request{URL: &url.URL{}},
					MarshalContent: func(v any) ([]byte, error) {
						switch v.(type) {
						case map[string]any:
							return nil, errm
						default:
							return json.Marshal(v)
						}
					},
				}
				defer test.Using(&nowUTC, func() time.Time { return time.Time{} })()
				defer test.Using(&makeErrorResponse, func(*Error, *Request) *Response {
					return &Response{ContentType: "Error-Content"}
				})()

				// ACT
				response := (&Problem{Status: http.StatusNotFound}).makeResponse(rq)

				// ASSERT
				test.That(t, response, "response").Equals(&Response{ContentType: "Error-Content"})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
