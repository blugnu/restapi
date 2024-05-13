package restapi

import (
	"testing"

	"github.com/blugnu/test"
)

func TestHeaders(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "set",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{
					"Canonical":       "old value (a)",
					"X-Canonicalised": "old value (b)",
				}}

				// ACT
				_ = sut.WithHeader("Canonical", "new value (a)")
				_ = sut.WithHeader("x-canonicalised", "new value (b)")
				_ = sut.WithHeader("new-header", "value (c)")

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Canonical":       "new value (a)",
					"X-Canonicalised": "new value (b)",
					"New-Header":      "value (c)",
				})
			},
		},
		{scenario: "setAll",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{
					"Canonical":     "old value (a)",
					"Canonicalised": "old value (b)",
				}}

				// ACT
				_ = sut.WithHeaders(map[string]any{
					"canonicalised": "new value (b)",
					"new-header":    "new value (c)",
				})

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"Canonical":     "old value (a)",
					"Canonicalised": "new value (b)",
					"New-Header":    "new value (c)",
				})
			},
		},
		{scenario: "setNonCanonical",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Result{headers: headers{}}

				// ACT
				_ = sut.WithNonCanonicalHeader("non-canonical", "value")

				// ASSERT
				test.Map(t, sut.headers).Equals(headers{
					"non-canonical": "value",
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
