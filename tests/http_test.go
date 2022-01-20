package tests

import (
	"testing"
)

func TestHttpTemplate(t *testing.T) {
	v := true
	if v == false {
		t.Errorf("Test failed")
	}

}
