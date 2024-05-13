package restapi

import (
	"encoding/xml"
	"fmt"
	"maps"
	"net/http"
	"sort"
	"time"
)

// function variables to facilitate testing
var xmlEncodeToken = func(e *xml.Encoder, t xml.Token) error {
	return e.EncodeToken(t)
}

// errorResponse provides the default projection model for a REST API Error.
//
// The struct is used to marshal an error response into the response body, supporting
// both JSON and XML marshalling.
//
// The struct is not exported, being used internally by the default implementation of
// the ProjectError function.  It is defined as a named struct type to facilitate
// unit tests of the restapi module.
//
// To change the representation of an error response applications should replace the
// implementation of the ProjectError function to return a value that can be marshalled
// to the content type (or types) required by that application.
type errorResponse struct {
	XMLName    xml.Name   `json:"-"`
	Status     int        `json:"status" xml:"status"`
	Error      string     `json:"error" xml:"error"`
	Message    string     `json:"message,omitempty" xml:"message,omitempty"`
	Path       string     `json:"path" xml:"path"`
	Query      string     `json:"query,omitempty" xml:"query,omitempty"`
	Timestamp  time.Time  `json:"timestamp" xml:"timestamp"`
	Help       string     `json:"help,omitempty" xml:"help,omitempty"`
	Additional errorProps `json:"additional,omitempty" xml:"additional,omitempty"`
}
type errorProps map[string]any

// MarshalXML marshals the map to XML, with each key:value pair mapped to
// a <key>value</key> element in the XML.
func (m errorProps) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}

	// sort the map keys to ensure consistent output (in terms of both
	// a stable key order for testing and also for consistency with
	// the JSON marshalling, which is also sorted by key)
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := m[key]
		t := xml.StartElement{Name: xml.Name{Local: key}}
		tokens = append(tokens, t, xml.CharData(fmt.Sprintf("%v", value)), xml.EndElement{Name: t.Name})
	}

	tokens = append(tokens, xml.EndElement{Name: start.Name})

	for _, t := range tokens {
		if err := xmlEncodeToken(e, t); err != nil {
			return err
		}
	}
	return e.Flush()
}

// ProjectError is called when writing an error response to obtain a representation
// of a REST API Error (the 'projection') to be used as the response body.  The function
// is a variable with a default implementation returning a struct with tags supporting
// both JSON and XML marshalling:
//
//	type struct {
//		XMLName    xml.Name       `json:"-"` // omit from JSON; set to "error" in XML
//		Status     int            `json:"status" xml:"status"`
//		Error      string         `json:"error" xml:"error"`
//		Message    string         `json:"message,omitempty" xml:"message,omitempty"`
//		Path       string         `json:"path" xml:"path"`
//		Query      string         `json:"query" xml:"query"`
//		Timestamp  time.Time      `json:"timestamp" xml:"timestamp"`
//		Help       string         `json:"help,omitempty" xml:"help,omitempty"`
//		Additional map[string]any `json:"additional,omitempty" xml:"additional,omitempty"`
//	}
//
// Applications may customise the body of error responses by replacing the implementation
// of this function and returning a custom struct or other type with marshalling support
// appropriate to the needs of the application.
var ProjectError = func(err ErrorInfo) any {
	// FUTURE: handling of []error, if present in the error (i.e. if implements Unwrap() []error)
	pe := errorResponse{
		XMLName:   xml.Name{Local: "error"},
		Status:    err.StatusCode,
		Error:     http.StatusText(err.StatusCode),
		Message:   err.Message,
		Path:      err.Request.URL.Path,
		Query:     err.Request.URL.RawQuery,
		Timestamp: err.TimeStamp,
		Help:      err.Help,
	}

	switch {
	case pe.Message == "" && err.Err != nil:
		pe.Message = err.Err.Error()

	case pe.Message != "" && err.Err != nil:
		pe.Message = err.Err.Error() + ": " + pe.Message
	}

	if len(err.Properties) > 0 {
		pe.Additional = maps.Clone(err.Properties)
	}

	return pe
}
