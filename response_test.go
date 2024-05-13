package restapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

func TestResponse(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "write/with headers",
			exec: func(t *testing.T) {
				// ARRANGE
				// using variables for the header keys avoids SA1008 issues from the static checker
				// (SA1008: warns against the use of non-canonical headers, which we are doing deliberately here)
				var ihdr = "int"
				var shdr = "string"
				rec := httptest.NewRecorder()
				sut := Response{
					StatusCode:  200,
					ContentType: "application/json",
					Content:     []byte("\"content\""),
					headers: map[string]any{
						ihdr: 42,
						shdr: "value",
					},
				}

				// ACT
				sut.write(rec, nil)

				// ASSERT
				result := rec.Result()
				test.That(t, result.StatusCode).Equals(200)
				test.That(t, result.Header.Get("Content-Type")).Equals("application/json")
				test.That(t, result.Header[ihdr]).Equals([]string{"42"})
				test.That(t, result.Header[shdr]).Equals([]string{"value"})
				test.That(t, rec.Body.String()).Equals("\"content\"")
			},
		},
		{scenario: "write/error writing response",
			exec: func(t *testing.T) {
				// ARRANGE
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}
				sut := Response{
					StatusCode:  200,
					ContentType: "application/json",
					Content:     []byte("\"content\""),
				}

				rwerr := errors.New("write error")
				defer test.Using(&responseWriterWrite, func(rw http.ResponseWriter, content []byte) error {
					return rwerr
				})()

				var loggedErr *InternalError
				defer test.Using(&LogError, func(info InternalError) {
					loggedErr = &info
				})()

				rq := &http.Request{URL: &url.URL{Path: "/path"}}

				// ACT
				sut.write(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(200)
				test.That(t, loggedErr.Err).Equals(rwerr)
				test.That(t, loggedErr.Message).Equals("error writing response")
				test.That(t, loggedErr.Help).Equals("(response: 200 OK): rw.Write() error: write error")
				test.That(t, loggedErr.Request).Equals(rq)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
