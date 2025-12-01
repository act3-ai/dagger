package internal

import (
	"testing"
)

func TestFoo(t *testing.T) {
	x := Foo(1)
	t.Log(x)
	if x != 14 {
		t.Fail()
	}
}
