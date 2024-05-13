package restapi

import (
	"testing"

	"github.com/blugnu/test"
)

func Test_nowUTC(t *testing.T) {
	// ACT
	result := nowUTC()

	// ASSERT
	test.IsFalse(t, result.IsZero())
	test.That(t, result.Location().String()).Equals("UTC")
}
