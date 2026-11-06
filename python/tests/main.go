// Testing module for python. All tests ran against testapp/ folder

package main

import (
	"context"
	"dagger/tests/internal/dagger"
	python "dagger/tests/internal/dagger/clients/python"
	"fmt"
	"strings"

	"dagger.io/dagger/core"
)

type Tests struct{}

// helper
func (t *Tests) srcDir() *dagger.Directory {
	src := core.NewCurrentModule().
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
	// d := dag.Directory().WithNewFile(scriptPath, script)
	out, err := python.New(t.srcDir()).
		Project().
		WithNewFile(scriptPath, script).
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
func (t *Tests) Mypy(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Mypy().Lint().Sync(ctx)

	return err
}

// +check
// Run pylint, expect valid/no errors
func (t *Tests) Pylint(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Pylint().Lint().Sync(ctx)

	return err
}

// +check
// Run pyright, expect valid/no errors
func (t *Tests) Pyright(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Pyright().Lint().Sync(ctx)

	return err

}

// +check
// Check if ruff version override works
func (t *Tests) RuffVersion(ctx context.Context) error {

	//add ruff to pyproject.toml first, then use that version to run lint
	pinnedVerDir := python.New(t.srcDir()).DevContainer().WithExec([]string{"uv", "add", "ruff==0.14.13"}).Directory("/app")

	pinnedVer, _ := python.New(pinnedVerDir).Ruff().Lint().WithExec([]string{"uv", "tree", "--package", "ruff"}).Stdout(ctx)

	if pinnedVer != "ruff v0.14.13\n" {
		return fmt.Errorf("Version expected: ruff v0.14.13, found: %s", pinnedVer)
	}

	return nil
}

// +check
// Run ruff lint, expect valid/no errors
func (t *Tests) RuffLint(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Ruff().Lint().Sync(ctx)

	return err
}

// +check
// Run ruff lint-report, expect valid/no errors
func (t *Tests) RuffLintReport(ctx context.Context) error {
	report := python.New(t.srcDir()).Ruff().LintReport()
	results, err := report.Contents(ctx)

	if err != nil {
		return err
	}

	if results != "[]" {
		return fmt.Errorf("Report found changes: %s", results)
	}

	return nil
}

// +check
// Run ruff lint-report with extra arguments, expect valid/no errors
func (t *Tests) RuffLintReportWithExtraArgs(ctx context.Context) error {
	expectedName := "ruff-lint.json"
	report := python.New(t.srcDir()).Ruff().LintReport(python.PythonRuffLintReportOpts{OutputFormat: "json", OutputFile: expectedName})
	results, err := report.Contents(ctx)

	if err != nil {
		return err
	}

	if results != "[]" {
		return fmt.Errorf("Report found changes: %s", results)
	}

	name, err := report.Name(ctx)

	if err != nil {
		return err
	}

	if name != expectedName {
		return fmt.Errorf("Report file incorrectly named. Expected '%s' got '%s'", expectedName, name)
	}

	return nil
}

// +check
// Run ruff lint-fix, expect valid/no errors
func (t *Tests) RuffLintFix(ctx context.Context) error {
	empty, err := python.New(t.srcDir()).Ruff().LintFix().IsEmpty(ctx)

	if !empty {
		return err
	}
	return nil
}

// +check
// Run ruff format, expect valid/no errors
func (t *Tests) RuffFormat(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Ruff().Format().Sync(ctx)

	return err
}

// +check
// Run ruff format-report, expect valid/no errors
func (t *Tests) RuffFormatReport(ctx context.Context) error {
	report := python.New(t.srcDir()).Ruff().FormatReport()
	results, err := report.Contents(ctx)

	if err != nil {
		return err
	}

	if results != "" {
		return fmt.Errorf("Report found changes: %s", results)
	}

	return nil
}

// +check
// Run ruff format-report with arguments, expect valid/no errors
func (t *Tests) RuffFormatReportWithExtraArgs(ctx context.Context) error {
	expectedName := "ruff-format.json"
	report := python.New(t.srcDir()).Ruff().FormatReport(python.PythonRuffFormatReportOpts{OutputFormat: "full", OutputFile: expectedName})
	results, err := report.Contents(ctx)

	if err != nil {
		return err
	}

	if results != "" {
		return fmt.Errorf("Report found changes: %s", results)
	}

	name, err := report.Name(ctx)

	if err != nil {
		return err
	}

	if name != expectedName {
		return fmt.Errorf("Report file incorrectly named. Expected '%s' got '%s'", expectedName, name)
	}

	return nil
}

// +check
// Run ruff format-fix, expect valid/no errors
func (t *Tests) RuffFormatFix(ctx context.Context) error {
	empty, err := python.New(t.srcDir()).Ruff().FormatFix().IsEmpty(ctx)

	if !empty {
		return err
	}
	return nil
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) Pytest(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Pytest().Test().Sync(ctx)

	return err
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) PytestReport(ctx context.Context) error {
	ptResults := python.New(t.srcDir()).Pytest().Report()

	checks := []struct {
		name string
		dir  *dagger.Directory
		file string
	}{
		{"xml coverage", ptResults.XML(), "coverage.xml"},
		{"xml junit", ptResults.XML(), "junit.xml"},
		{"json coverage", ptResults.JSON(), "coverage.json"},
		{"html index", ptResults.HTML(), "index.html"},
	}

	for _, c := range checks {
		exists, err := c.dir.Exists(ctx, c.file)
		if err != nil {
			return fmt.Errorf("checking %s: %w", c.name, err)
		}
		if !exists {
			return fmt.Errorf("%s not found (%s)", c.name, c.file)
		}
	}

	return nil
}

// +check
// Run cognitive complexity lint, expect valid/no errors
func (t *Tests) CognitiveComplexity(ctx context.Context) error {
	_, err := python.New(t.srcDir()).CognitiveComplexity().Lint().Sync(ctx)

	return err
}

// +check
// Run ty, expect valid/no errors
func (t *Tests) Ty(ctx context.Context) error {
	_, err := python.New(t.srcDir()).Ty().Lint().Sync(ctx)

	return err
}

// +check
// Run ty, expect valid/no errors
func (t *Tests) TyReport(ctx context.Context) error {
	report := python.New(t.srcDir()).Ty().Report()
	results, err := report.Contents(ctx)

	if err != nil {
		return err
	}

	if results != "[]" {
		return fmt.Errorf("Report found changes: %s", results)
	}

	return nil
}

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) UvProject(ctx context.Context) error {
	src := core.NewDirectory().
		WithDirectory("sub", t.srcDir())

	_, err := python.New(src, python.PythonOpts{
		SyncArgs:  []string{"--project", "/app/sub"},
		SyncPaths: []string{"sub/pyproject.toml", "sub/uv.lock"},
	}).
		DevContainer().Sync(ctx)

	return err
}
