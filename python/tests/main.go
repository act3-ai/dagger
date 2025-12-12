// Testing module for python. All tests ran against testapp/ folder

package main

import (
	"context"
	"dagger/tests/internal/dagger"
)

type Tests struct{}

// helper
func (t *Tests) srcDir() *dagger.Directory {
	src := dag.CurrentModule().
		Source().
		Directory("testapp").
		Filter(dagger.DirectoryFilterOpts{Exclude: []string{"err.py"}})

	return src
}

// +check
// Run mypy, expect valid/no errors
func (t *Tests) Mypy(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Mypy().Check(ctx)
}

// +check
// Run pylint, expect valid/no errors
func (t *Tests) Pylint(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Pylint().Check(ctx)
}

// +check
// Run pyright, expect valid/no errors
func (t *Tests) Pyright(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Pyright().Check(ctx)

}

// +check
// Run ruff lint, expect valid/no errors
func (t *Tests) RuffLint(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Ruff().Lint().Check(ctx)
}

// +check
// Run ruff-format, expect valid/no errors
func (t *Tests) RuffFormat(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Ruff().Format().Check(ctx)
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) Pytest(ctx context.Context,
) error {
	return dag.Python(t.srcDir()).Pytest().Check(ctx)
}
