// Testing module for python. All tests ran against testapp/ folder

package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"fmt"
	"strings"
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

const script = `# /// script
# requires-python = ">=3.13"
# dependencies = ["six"]
# ///

print("it worked")
`

// +check
// Run mypy, expect valid/no errors
func (t *Tests) Base(ctx context.Context) error {
	const scriptPath = "script.py"
	d := dag.Directory().WithNewFile(scriptPath, script)
	out, err := dag.
		Python(d).
		Base().
		WithExec([]string{"uv", "run", scriptPath}).
		Stdout(ctx)
	if err != nil {
		return err
	}
	out = strings.TrimSpace(out)
	if out != "it worked" {
		return fmt.Errorf("expected \"it worked\" but got %q", out)
	}
	return err
}

// +check
// Run mypy, expect valid/no errors
func (t *Tests) Mypy() *dagger.Container {
	return dag.Python(t.srcDir()).Mypy().Lint()
}

// +check
// Run pylint, expect valid/no errors
func (t *Tests) Pylint() *dagger.Container {
	return dag.Python(t.srcDir()).Pylint().Lint()
}

// +check
// Run pyright, expect valid/no errors
func (t *Tests) Pyright() *dagger.Container {
	return dag.Python(t.srcDir()).Pyright().Lint()

}

// +check
// Run ruff lint, expect valid/no errors
func (t *Tests) RuffLint() *dagger.Container {
	return dag.Python(t.srcDir()).Ruff().Lint()
}

// +check
// Run ruff-format, expect valid/no errors
func (t *Tests) RuffFormat(ctx context.Context) error {
	empty, err := dag.Python(t.srcDir()).Ruff().Fix().IsEmpty(ctx)

	if !empty {
		return err
	}
	return nil
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) Pytest() *dagger.Container {
	return dag.Python(t.srcDir()).Pytest().Test()
}
