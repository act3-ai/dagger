// Testing module for python. All tests ran against testapp/ folder

package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"encoding/json"
	"fmt"
)

type Tests struct{}

// helper
func (t *Tests) srcDir(ctx context.Context,
	// +optional
	exclude []string,
) *dagger.Directory {
	src := dag.CurrentModule().
		Source().
		Directory("testapp").
		Filter(dagger.DirectoryFilterOpts{Exclude: exclude})

	return src
}

// helper
func (t *Tests) checkExitCode(ctx context.Context,
	name string,
	exitCode int,
	results *dagger.File,
) error {
	if exitCode == 0 {
		return nil
	}

	out, err := results.Contents(ctx)
	if err != nil {
		return err
	}

	return fmt.Errorf("%s failed: %s", name, out)
}

// +check
// Run mypy, expect valid/no errors
func (t *Tests) Mypy(ctx context.Context,
) error {
	src := t.srcDir(ctx, []string{"err.py"})
	mypy := dag.Python(src).Mypy()
	exitCode, err := mypy.ExitCode(ctx)
	if err != nil {
		return err
	}

	return t.checkExitCode(ctx, "mypy", exitCode, mypy.Results())
}

// +check
// Run pylint, expect valid/no errors
func (t *Tests) Pylint(ctx context.Context,
) error {
	src := t.srcDir(ctx, []string{"err.py"})
	pylint := dag.Python(src).Pylint()
	exitCode, err := pylint.ExitCode(ctx)
	if err != nil {
		return err
	}

	return t.checkExitCode(ctx, "pylint", exitCode, pylint.Results())
}

// +check
// Run pylint, expect valid/no errors
func (t *Tests) Pyright(ctx context.Context,
) error {
	src := t.srcDir(ctx, []string{"err.py"})
	pyright := dag.Python(src).Pyright()
	exitCode, err := pyright.ExitCode(ctx)
	if err != nil {
		return err
	}

	return t.checkExitCode(ctx, "pyright", exitCode, pyright.Results())
}

// +check
// Run ruff-check, expect valid/no errors
func (t *Tests) RuffCheck(ctx context.Context,
) error {
	src := t.srcDir(ctx, []string{"err.py"})
	ruffCheck := dag.Python(src).Pyright()
	exitCode, err := ruffCheck.ExitCode(ctx)
	if err != nil {
		return err
	}

	return t.checkExitCode(ctx, "ruff-check", exitCode, ruffCheck.Results())
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) UnitTest(ctx context.Context,
) error {
	src := t.srcDir(ctx, []string{"err.py"})
	unitTest := dag.Python(src).UnitTest()
	jsonResults, err := unitTest.JSON().Contents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get file contents: %s", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonResults), &data); err != nil {
		return fmt.Errorf("failed to parse json: %s", err)
	}

	totals, ok := data["totals"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to parse totals: %s", err)
	}

	pct, ok := totals["percent_covered"].(float64)
	if !ok {
		return fmt.Errorf("failed to parse percent_covered: %s", err)
	}

	if pct < 100.0 {
		return fmt.Errorf("Code coverage not at 100%: %s", pct)
	}

	return nil
}
