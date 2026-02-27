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
// Check if ruff version override works
func (t *Tests) RuffVersion(ctx context.Context) error {

	//add ruff to pyproject.toml first, then use that version to run lint
	pinnedVerDir := dag.Python(t.srcDir()).Base().WithExec([]string{"uv", "add", "ruff==0.14.13"}).Directory("/app")

	pinnedVer, _ := dag.Python(pinnedVerDir).Ruff().Lint().WithExec([]string{"uv", "tree", "--package", "ruff"}).Stdout(ctx)

	if pinnedVer != "ruff v0.14.13\n" {
		return fmt.Errorf("Version expected: ruff v0.14.13, found: %s", pinnedVer)
	}

	return nil
}

// +check
// Run ruff lint, expect valid/no errors
func (t *Tests) RuffLint() *dagger.Container {
	return dag.Python(t.srcDir()).Ruff().Lint()
}

// +check
// Run ruff lint-report, expect valid/no errors
func (t *Tests) RuffLintReport(ctx context.Context) error {
	report := dag.Python(t.srcDir()).Ruff().LintReport()
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
	report := dag.Python(t.srcDir()).Ruff().LintReport(dagger.PythonRuffLintReportOpts{OutputFormat: "json", OutputFile: expectedName})
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
	empty, err := dag.Python(t.srcDir()).Ruff().LintFix().IsEmpty(ctx)

	if !empty {
		return err
	}
	return nil
}

// +check
// Run ruff format, expect valid/no errors
func (t *Tests) RuffFormat() *dagger.Container {
	return dag.Python(t.srcDir()).Ruff().Format()
}

// +check
// Run ruff format-report, expect valid/no errors
func (t *Tests) RuffFormatReport(ctx context.Context) error {
	report := dag.Python(t.srcDir()).Ruff().FormatReport()
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
	report := dag.Python(t.srcDir()).Ruff().FormatReport(dagger.PythonRuffFormatReportOpts{OutputFormat: "full", OutputFile: expectedName})
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
	empty, err := dag.Python(t.srcDir()).Ruff().FormatFix().IsEmpty(ctx)

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

// +check
// Run unit-test, expect valid/no errors
func (t *Tests) PytestReport(ctx context.Context) error {
	ptResults := dag.Python(t.srcDir()).Pytest().Report()

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
// Run flake8 lint, expect valid/no errors
func (t *Tests) CognitiveComplexity() *dagger.Container {
	return dag.Python(t.srcDir()).CognitiveComplexity().Lint()
}
