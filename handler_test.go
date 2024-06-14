package restapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

type Recorder struct {
	*httptest.ResponseRecorder
	statusCode int
}

func (r *Recorder) Content() string {
	return r.ResponseRecorder.Body.String()
}

func (r *Recorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseRecorder.WriteHeader(statusCode)
}

type fakeHandler struct {
	isCalled bool
}

func (h *fakeHandler) ServeAPI(_ context.Context, _ *http.Request) any {
	h.isCalled = true
	return nil
}

func TestHandler(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "EndpointFunc",
			exec: func(t *testing.T) {
				// ARRANGE
				isCalled := false
				fn := func(_ context.Context, _ *http.Request) any {
					isCalled = true
					return nil
				}

				// ACT
				_ = EndpointFunc(fn).ServeAPI(context.Background(), nil)

				// ASSERT
				test.IsTrue(t, isCalled)
			},
		},
		{scenario: "Handler",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{}
				rec := httptest.NewRecorder()
				h := &fakeHandler{}

				// ACT
				Handler(h)(rec, rq)

				// ASSERT
				test.IsTrue(t, h.isCalled)
			},
		},
		{scenario: "HandlerFunc/invalid request Accept header",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{Header: http.Header{"Accept": []string{"text/plain"}}}
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}
				isLogged := false
				defer test.Using(&LogError, func(InternalError) {
					isLogged = true
				})()

				// ACT
				HandlerFunc(func(_ context.Context, rq *http.Request) any {
					return nil
				})(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(http.StatusNotAcceptable)
				test.That(t, rec.Content()).Equals("[\"application/json\",\"application/xml\",\"text/json\",\"test/xml\",\"*/*\",none]")
				test.IsTrue(t, isLogged)
			},
		},
		{scenario: "HandlerFunc/other request error",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{}
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}

				rqerr := errors.New("request error")
				defer test.Using(&newRequest, func(*http.Request) (*Request, error) {
					return nil, rqerr
				})()

				isLogged := false
				defer test.Using(&LogError, func(InternalError) {
					isLogged = true
				})()

				// ACT
				HandlerFunc(func(_ context.Context, rq *http.Request) any {
					return nil
				})(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(http.StatusInternalServerError)
				test.That(t, rec.Content()).Equals("\"request error\"")
				test.IsTrue(t, isLogged)
			},
		},
		{scenario: "HandlerFunc/other request error/error writing response",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{}
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}

				rqerr := errors.New("request error")
				defer test.Using(&newRequest, func(*http.Request) (*Request, error) {
					return nil, rqerr
				})()

				rwerr := errors.New("response writer error")
				defer test.Using(&responseWriterWrite, func(rw http.ResponseWriter, content []byte) error {
					return rwerr
				})()

				logged := []InternalError{}
				defer test.Using(&LogError, func(inf InternalError) {
					logged = append(logged, inf)
				})()

				// ACT
				HandlerFunc(func(_ context.Context, rq *http.Request) any {
					return nil
				})(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(http.StatusInternalServerError)
				test.That(t, len(logged)).Equals(2)
				test.That(t, logged[0].Err).Equals(rqerr)
				test.That(t, logged[1].Err).Equals(rwerr)
				test.That(t, logged[1].Message).Equals("error writing request error response")
				test.That(t, logged[1].Help).Equals("(request error: request error): rw.Write() error: response writer error")
			},
		},
		{scenario: "HandlerFunc/panic",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{URL: &url.URL{Path: "/path"}}
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}

				logged := []InternalError{}
				defer test.Using(&LogError, func(inf InternalError) {
					logged = append(logged, inf)
				})()

				// ACT
				HandlerFunc(func(_ context.Context, rq *http.Request) any {
					panic("panic")
				})(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(http.StatusInternalServerError)
				test.That(t, len(logged)).Equals(1)
				test.That(t, logged[0].Message).Equals("handler panic")
				test.That(t, logged[0].Request).Equals(rq)
			},
		},
		{scenario: "HandlerFunc/successful",
			exec: func(t *testing.T) {
				// ARRANGE
				rq := &http.Request{}
				rec := &Recorder{ResponseRecorder: httptest.NewRecorder()}

				defer test.Using(&newRequest, func(*http.Request) (*Request, error) {
					return &Request{}, nil
				})()

				// ACT
				HandlerFunc(func(_ context.Context, rq *http.Request) any {
					return http.StatusOK
				})(rec, rq)

				// ASSERT
				test.That(t, rec.statusCode).Equals(http.StatusOK)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE

			// ACT
			tc.exec(t)
		})
	}
}
