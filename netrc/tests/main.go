// A generated module for Tests functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
)

type Tests struct{}

const (
	expected = `machine myreg.com
login myuser
password MyPass1
machine myreg2.com
login myuser2
password MyPass1
`
)

// ensures netrc matches output of secret
// +check
func (m *Tests) Secret(ctx context.Context) error {

	pw := dag.SetSecret("MY_PW", "MyPass1")

	mySecret := dag.Netrc().
		WithLogin("myreg.com", "myuser", pw).
		WithLogin("myreg2.com", "myuser2", pw).
		AsSecret()

	out, err := mySecret.Plaintext(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute")
	}

	if out != expected {
		return fmt.Errorf("output does not match\nexpected:\n%s \nactual:\n%s", expected, out)
	}

	return nil
}
