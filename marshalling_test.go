package restapi

import (
	"encoding/xml"
	"testing"

	"github.com/blugnu/test"
)

func TestMarshalling(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "application/json",
			exec: func(t *testing.T) {
				// ACT
				result, _ := marshal["application/json"](struct{ A int }{A: 1})

				// ASSERT
				test.That(t, string(result)).Equals(`{"A":1}`)
			},
		},
		{scenario: "application/xml",
			exec: func(t *testing.T) {
				// ACT
				result, _ := marshal["application/xml"](struct {
					XMLName xml.Name
					A       int
				}{
					XMLName: xml.Name{Local: "struct"},
					A:       1,
				})

				// ASSERT
				test.That(t, string(result)).Equals(`<struct><A>1</A></struct>`)
			},
		},
		{scenario: "text/json",
			exec: func(t *testing.T) {
				// ACT
				result, _ := marshal["text/json"](struct{ A int }{A: 1})

				// ASSERT
				test.That(t, string(result)).Equals("{\n  \"A\": 1\n}")
			},
		},
		{scenario: "text/xml",
			exec: func(t *testing.T) {
				// ACT
				result, _ := marshal["text/xml"](struct {
					XMLName xml.Name
					A       int
				}{
					XMLName: xml.Name{Local: "struct"},
					A:       1,
				})

				// ASSERT
				test.That(t, string(result)).Equals("<struct>\n    <A>1</A>\n</struct>")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
