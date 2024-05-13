package restapi

import (
	"encoding/json"
	"encoding/xml"
)

type marshalFunc = func(any) ([]byte, error)

// marshal is a map that associates each supported Content-Type to an
// appropriate marshalling function.
//
// FUTURE: more complete xml support (charset, etc.)
// FUTURE: support q weights
// FUTURE: additional content types (e.g. yaml)
// FUTURE: configurable content types and marshallers
var marshal = map[string]marshalFunc{
	"application/json": json.Marshal,
	"application/xml":  xml.Marshal,
	"text/json":        func(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") },
	"text/xml":         func(v any) ([]byte, error) { return xml.MarshalIndent(v, "", "    ") },
}
