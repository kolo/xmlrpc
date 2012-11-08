package xmlrpc

import (
	"testing"
)

func assert_not_nil(t *testing.T, object interface{}) {
	if object == nil {
		t.Error("Expected object not to be a nil, but it is.")
	}
}

func assert_nil(t *testing.T, object interface{}) {
	if object != nil {
		t.Error("Expected object to be a nil, but it is not.")
	}
}

func assert_equal(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Error("Expected objects to be equal, but they are not.")
	}
}
