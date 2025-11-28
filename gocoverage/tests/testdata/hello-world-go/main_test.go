package main

import "testing"

func TestFoo(t *testing.T) {
	x := foo(1)
	t.Log(x)
	if x != 14 {
		t.Fail()
	}
}
