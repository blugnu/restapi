package restapi

import (
	"testing"

	"github.com/blugnu/test"
)

func TestLogErr(t *testing.T) {
	// LogError is a NO-OP function by default so the only thing to
	// test is that it does not panic when called
	defer test.ExpectPanic(nil).Assert(t)

	// ACT
	LogError(InternalError{})
}
