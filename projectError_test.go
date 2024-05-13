package restapi

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"maps"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestProjectError(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "message only",
			exec: func(t *testing.T) {
				// ARRANGE
				ts := time.Date(2012, 11, 10, 9, 8, 7, 0, time.UTC)
				inf := ErrorInfo{
					StatusCode: 400,
					Request: &http.Request{URL: &url.URL{
						Path:     "/api/v1/test",
						RawQuery: "param=value",
					}},
					Message:   "message",
					TimeStamp: ts,
				}

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				test.That(t, projection, "projection").Equals(errorResponse{
					XMLName:   xml.Name{Local: "error"},
					Status:    400,
					Error:     "Bad Request",
					Message:   "message",
					Path:      "/api/v1/test",
					Query:     "param=value",
					Timestamp: ts,
				})
			},
		},
		{scenario: "error only",
			exec: func(t *testing.T) {
				// ARRANGE
				ts := time.Date(2012, 11, 10, 9, 8, 7, 0, time.UTC)
				e := errors.New("error")
				inf := ErrorInfo{
					StatusCode: 400,
					Request: &http.Request{URL: &url.URL{
						Path:     "/api/v1/test",
						RawQuery: "param=value",
					}},
					Err:       e,
					TimeStamp: ts,
				}

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				test.That(t, projection, "projection").Equals(errorResponse{
					XMLName:   xml.Name{Local: "error"},
					Status:    400,
					Error:     "Bad Request",
					Message:   "error",
					Path:      "/api/v1/test",
					Query:     "param=value",
					Timestamp: ts,
				})
			},
		},
		{scenario: "message and error",
			exec: func(t *testing.T) {
				// ARRANGE
				ts := time.Date(2012, 11, 10, 9, 8, 7, 0, time.UTC)
				e := errors.New("error")
				inf := ErrorInfo{
					StatusCode: 400,
					Request: &http.Request{URL: &url.URL{
						Path:     "/api/v1/test",
						RawQuery: "param=value",
					}},
					Err:       e,
					Message:   "message",
					TimeStamp: ts,
				}

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				test.That(t, projection, "projection").Equals(errorResponse{
					XMLName:   xml.Name{Local: "error"},
					Status:    400,
					Error:     "Bad Request",
					Message:   "error: message",
					Path:      "/api/v1/test",
					Query:     "param=value",
					Timestamp: ts,
				})
			},
		},

		{scenario: "no properties",
			exec: func(t *testing.T) {
				// ARRANGE
				ts := time.Date(2012, 11, 10, 9, 8, 7, 0, time.UTC)
				e := errors.New("oops")
				inf := ErrorInfo{
					StatusCode: 400,
					Err:        e,
					Request: &http.Request{URL: &url.URL{
						Path:     "/api/v1/test",
						RawQuery: "param=value",
					}},
					Help:      "you need some",
					TimeStamp: ts,
				}

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				test.That(t, projection, "projection").Equals(errorResponse{
					XMLName:   xml.Name{Local: "error"},
					Status:    400,
					Error:     "Bad Request",
					Message:   "oops",
					Path:      "/api/v1/test",
					Query:     "param=value",
					Timestamp: ts,
					Help:      "you need some",
				})

				// ASSERT
				t.Run("json", func(t *testing.T) {
					result, err := json.Marshal(projection)
					test.Error(t, err).IsNil()
					test.String(t, result).Equals(`{` +
						`"status":400,` +
						`"error":"Bad Request",` +
						`"message":"oops",` +
						`"path":"/api/v1/test",` +
						`"query":"param=value",` +
						`"timestamp":"2012-11-10T09:08:07Z",` +
						`"help":"you need some"` +
						`}`,
					)
				})

				t.Run("xml", func(t *testing.T) {
					result, err := xml.Marshal(projection)
					test.Error(t, err).IsNil()
					test.String(t, result).Equals(`<error>` +
						`<status>400</status>` +
						`<error>Bad Request</error>` +
						`<message>oops</message>` +
						`<path>/api/v1/test</path>` +
						`<query>param=value</query>` +
						`<timestamp>2012-11-10T09:08:07Z</timestamp>` +
						`<help>you need some</help>` +
						`</error>`,
					)
				})
			},
		},
		{scenario: "with properties",
			exec: func(t *testing.T) {
				// ARRANGE
				ts := time.Date(2012, 11, 10, 9, 8, 7, 0, time.UTC)
				e := errors.New("oops")
				inf := ErrorInfo{
					StatusCode: 400,
					Err:        e,
					Request: &http.Request{URL: &url.URL{
						Path:     "/api/v1/test",
						RawQuery: "param=value",
					}},
					Help:      "you need some",
					TimeStamp: ts,
					Properties: map[string]any{
						"skey": "value",
						"ikey": 42,
					},
				}

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				test.That(t, projection.(errorResponse), "projection").Equals(errorResponse{
					XMLName:   xml.Name{Local: "error"},
					Status:    400,
					Error:     "Bad Request",
					Message:   "oops",
					Path:      "/api/v1/test",
					Query:     "param=value",
					Timestamp: ts,
					Help:      "you need some",
					Additional: map[string]any{
						"ikey": 42,
						"skey": "value",
					},
				}, func(got, wanted errorResponse) bool {
					gprops := got.Additional
					wprops := wanted.Additional
					got.Additional = nil
					wanted.Additional = nil
					return reflect.DeepEqual(got, wanted) && maps.Equal(gprops, wprops)
				})

				// ASSERT
				t.Run("json", func(t *testing.T) {
					result, err := json.Marshal(projection)
					test.Error(t, err).IsNil()
					test.String(t, result).Equals(`{` +
						`"status":400,` +
						`"error":"Bad Request",` +
						`"message":"oops",` +
						`"path":"/api/v1/test",` +
						`"query":"param=value",` +
						`"timestamp":"2012-11-10T09:08:07Z",` +
						`"help":"you need some",` +
						`"additional":{` +
						`"ikey":42,` +
						`"skey":"value"` +
						`}` +
						`}`,
					)
				})

				t.Run("xml", func(t *testing.T) {
					result, err := xml.Marshal(projection)
					test.Error(t, err).IsNil()
					test.String(t, result).Equals(`<error>` +
						`<status>400</status>` +
						`<error>Bad Request</error>` +
						`<message>oops</message>` +
						`<path>/api/v1/test</path>` +
						`<query>param=value</query>` +
						`<timestamp>2012-11-10T09:08:07Z</timestamp>` +
						`<help>you need some</help>` +
						`<additional>` +
						`<ikey>42</ikey>` +
						`<skey>value</skey>` +
						`</additional>` +
						`</error>`,
					)
				})
			},
		},
		{scenario: "with properties/xml encoding error",
			exec: func(t *testing.T) {
				// ARRANGE
				inf := ErrorInfo{
					Request: &http.Request{URL: &url.URL{}},
					// we need at least one property to trigger the error
					Properties: map[string]any{
						"key": "value",
					},
				}
				xmlerr := errors.New("oops")
				defer test.Using(&xmlEncodeToken, func(e *xml.Encoder, t xml.Token) error { return xmlerr })()

				// ACT
				projection := ProjectError(inf)

				// ASSERT
				_, err := xml.Marshal(projection)
				test.Error(t, err).Is(xmlerr)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
