// Testing module for python. All tests ran against testapp/ folder

package main

import (
	"context"
	"dagger/tests/internal/dagger"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// Run all tests
func (t *Tests) All(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {
	p := pool.New().WithErrors().WithContext(ctx)

	errDir := dag.Directory().WithDirectory(".", src)
	validDir := dag.Directory().WithDirectory(".", src, dagger.DirectoryWithDirectoryOpts{Exclude: []string{"err.py"}})

	// mypy tests
	p.Go(func(ctx context.Context) error {
		return t.Mypy(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.MypyIgnoreErr(ctx, errDir)
	})

	// pylint tests
	p.Go(func(ctx context.Context) error {
		return t.Pylint(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.PylintIgnoreErr(ctx, errDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.PylintOutputJson(ctx, validDir)
	})

	// pyright tests
	p.Go(func(ctx context.Context) error {
		return t.Pyright(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.PyrightIgnoreErr(ctx, errDir)
	})

	// ruff tests
	p.Go(func(ctx context.Context) error {
		return t.Ruffcheck(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.RuffcheckIgnoreErr(ctx, errDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.RuffcheckOutputJson(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.RuffFormat(ctx, validDir)
	})
	p.Go(func(ctx context.Context) error {
		return t.RuffFormatIgnoreErr(ctx, errDir)
	})

	// unit test
	p.Go(func(ctx context.Context) error {
		return t.UnitTest(ctx, validDir)
	})
	return p.Wait()
}

// Run mypy, expect valid/no errors
func (t *Tests) Mypy(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	_, err := dag.Python(src).Mypy(ctx)

	return err
}

// Run mypy, expect lint err but dagger still pass
func (t *Tests) MypyIgnoreErr(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	_, err := dag.Python(src).Mypy(ctx, dagger.PythonMypyOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}

	return err
}

// Run pylint, expect valid/no errors
func (t *Tests) Pylint(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	_, err := dag.Python(src).Pylint(ctx)

	return err
}

// Run pylint, expect lint err but dagger still pass
func (t *Tests) PylintIgnoreErr(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	_, err := dag.Python(src).Pylint(ctx, dagger.PythonPylintOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}

	return err
}

// Run pylint, output json and expect valid
func (t *Tests) PylintOutputJson(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	out, err := dag.Python(src).Pylint(ctx, dagger.PythonPylintOpts{OutputFormat: "json"})

	if !json.Valid([]byte(out)) {
		return fmt.Errorf("Invalid Json")
	}

	return err
}

// Run pyright, expect valid/no errors
func (t *Tests) Pyright(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	_, err := dag.Python(src).Pyright(ctx)

	return err
}

// Run pyright, expect lint err but dagger still pass
func (t *Tests) PyrightIgnoreErr(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	_, err := dag.Python(src).Pyright(ctx, dagger.PythonPyrightOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}

	return err
}

// Run ruffcheck expect valid/no errors
func (t *Tests) Ruffcheck(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	_, err := dag.Python(src).RuffCheck(ctx)

	return err
}

// Run ruffcheck expect lint err but dagger still pass
func (t *Tests) RuffcheckIgnoreErr(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	_, err := dag.Python(src).RuffCheck(ctx, dagger.PythonRuffCheckOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}

	return err
}

// Run ruffcheck output json and expect valid
func (t *Tests) RuffcheckOutputJson(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	out, err := dag.Python(src).RuffCheck(ctx, dagger.PythonRuffCheckOpts{OutputFormat: "json"})

	if !json.Valid([]byte(out)) {
		return fmt.Errorf("Invalid Json")
	}
	return err
}

// Run RuffFormat expect lint err but dagger still pass
func (t *Tests) RuffFormat(ctx context.Context,
	// +defaultPath="./testapp"
	// +ignore=["err.py"]
	src *dagger.Directory) error {

	_, err := dag.Python(src).RuffFormat(ctx)

	return err
}

// Run RuffFormat and expect valid
func (t *Tests) RuffFormatIgnoreErr(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	_, err := dag.Python(src).RuffFormat(ctx, dagger.PythonRuffFormatOpts{IgnoreError: true})
	if err != nil {
		return fmt.Errorf("failed to ignore exec errors: %w", err)
	}

	return err
}

// Run UnitTest and ensure results/ directory contains expected contents/files
func (t *Tests) UnitTest(ctx context.Context,
	// +defaultPath="./testapp"
	src *dagger.Directory) error {

	//expected contents of results/ directory
	expected := []string{"html/", "pytest-junit.xml", "unit-test.xml"}

	actual, err := dag.Python(src).UnitTest().Entries(ctx)

	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("actual %v, expected %v", actual, expected)
	}

	return err
}
