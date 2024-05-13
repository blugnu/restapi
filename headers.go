package restapi

import "net/http"

// headers is a map of response headers; by using this type
// rather than explicit map declaration the headers field in the
// apiresult struct is autoinitialised  i.e. and we don't have to
// worry about nil map errors
type headers map[string]any

// WithHeader sets a header to be included in the response for the error.
//
// The specified header will be added to any headers already set on the Error.
// If the specified header is already set on the Error the existing header will
// be replaced with the new value.
//
// The header key is canonicalised using http.CanonicalHeaderKey.  To set a header
// with a non-canonical key use WithNonCanonicalHeader.
func (h headers) set(k string, v any) {
	h[http.CanonicalHeaderKey(k)] = v
}

// WithHeaders sets the headers to be included in the response for the error.
//
// The specified headers will be added to any headers already set on the Error.
// If the new headers contain values already set on the Error the existing headers
// will be replaced with the new values.
//
// The header keys are canonicalised using http.CanonicalHeaderKey.  To set a header
// with a non-canonical key use WithNonCanonicalHeader.
func (h headers) setAll(headers map[string]any) {
	for k, v := range headers {
		h[http.CanonicalHeaderKey(k)] = v
	}
}

// WithNonCanonicalHeader sets a non-canonical header to be included in the response
// for the error.
//
// The specified header will be added to any headers already set on the Error.
// If the specified header is already set on the Error the existing header will
// be replaced with the new value.
//
// The header key is not canonicalised; if the specified key is canonical then the
// canonical header will be set.
//
// WithNonCanonicalHeader should only be used when a non-canonical header key is
// specifically required (which is rare).  Ordinarily WithHeader should be used.
func (h headers) setNonCanonical(k string, v any) {
	h[k] = v
}
