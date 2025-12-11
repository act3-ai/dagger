// Markdownlint provides utilities for running markdownlint-cli2 without installing locally with npm, brew, or docker.
// See https://github.com/DavidAnson/markdownlint-cli2 for more info. Most configuration should be done in '.markdownlint-cli2.yaml'
// within the source directory, or provided using 'WithConfig'.

package main

import (
	"context"
	"dagger/markdownlint/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "docker.io/davidanson/markdownlint-cli2"

type Markdownlint struct {
	Base *dagger.Container

	// +private
	Command []string

	// +private
	disableDefaultGlobs bool
}

type MarkdownLintResults struct {
	// prints the combined output of stdout and stderr as a string
	// +private
	Output string
	// returns results of markdownlint-cli2 as a file
	Results *dagger.File
	// returns exit code of markdownlint-cli2
	ExitCode int
}

type MarkdownLintAutoFixResults struct {
	// returns results of markdownlint autofix as a changeset
	Changes *dagger.Changeset
}

func New(ctx context.Context,
	// Source directory containing markdown files to be linted.
	// +ignore=["**", "!**/*.md", "!.markdownlint*", "!package.json"]
	src *dagger.Directory,

	// Custom container to use as a base container. Must have 'markdownlint-cli2' available on PATH.
	// +optional
	base *dagger.Container,

	// Version (image tag) to use as a markdownlint-cli2 binary source.
	// +optional
	// +default="latest"
	version string,

	// Configuration file.
	// +optional
	config *dagger.File,
) *Markdownlint {
	if base == nil {
		base = dag.Container().
			From(fmt.Sprintf("%s:%s", defaultImageRepository, version)).
			WithUser("root")
	}

	var disableDefaultGlobs bool
	cfgFiles, err := src.Filter(dagger.DirectoryFilterOpts{Include: []string{".markdownlint*"}}).
		Entries(ctx)
	if err != nil {
		panic(fmt.Errorf("discovering config files: %w", err))
	}
	if len(cfgFiles) > 0 {
		disableDefaultGlobs = true
	}

	cmd := []string{"markdownlint-cli2"}
	srcDir := "/work/src"
	base = base.With(
		func(c *dagger.Container) *dagger.Container {
			if config != nil {
				cfgPath, err := config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, config)
				cmd = append(cmd, "--config", cfgPath)
				disableDefaultGlobs = true
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	return &Markdownlint{
		Base:                base,
		Command:             cmd,
		disableDefaultGlobs: disableDefaultGlobs,
	}
}

// Runs markdownlint-cli2 against a given source directory. Returns a results file and an exit-code.
func (m *Markdownlint) Lint(ctx context.Context,
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,
) (*MarkdownLintResults, error) {
	cmd := m.Command
	cmd = append(cmd, extraArgs...)

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}

	ctr, err := m.Base.WithExec(cmd, dagger.ContainerWithExecOpts{
		Expect: dagger.ReturnTypeAny}).Sync(ctx)

	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running markdownlint-cli2: %w", err)
	}
	output, err := ctr.CombinedOutput(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("getting output: %w", err)
	}

	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		// exit code not found
		return nil, fmt.Errorf("get exit code: %w", err)
	}

	return &MarkdownLintResults{
		Output:   output,
		Results:  dag.File("markdownlint-results.txt", output),
		ExitCode: exitCode,
	}, nil
}

// Check for any errors running markdownlint-cli2
func (ml *MarkdownLintResults) Check(ctx context.Context) error {
	if ml.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("%s", ml.Output)
}

// AutoFix attempts to fix any linting errors reported by rules that emit fix information.
// Returns a Changeset that can be used to apply any changes made
// to the host.
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) AutoFix(ctx context.Context,
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string) (*MarkdownLintAutoFixResults, error) {
	cmd := m.Command
	cmd = append(cmd, "--fix")

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}
	ctr, err := m.Base.WithUser("root").
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny}).Sync(ctx)
	if err != nil {
		// unexpected error
		return nil, fmt.Errorf("running markdownlint autofix: %w", err)
	}

	afterChanges := ctr.Directory("/work/src").Filter(dagger.DirectoryFilterOpts{Exclude: []string{""}})

	return &MarkdownLintAutoFixResults{
		Changes: afterChanges.Changes(m.Base.Directory("/work/src")),
	}, nil
}

// returns the results of markdownlint autofix as a changeset that can be applied to the host.
func (mr *MarkdownLintAutoFixResults) Fix() (*dagger.Changeset, error) {
	return mr.Changes, nil
}

// Returns an error if markdownlint autofix made any changes
func (mr *MarkdownLintAutoFixResults) Check(ctx context.Context) error {
	empty, err := mr.Changes.IsEmpty(ctx)
	if err != nil {
		return err
	}

	if empty {
		return nil
	}

	diff, err := mr.Changes.AsPatch().Contents(ctx)
	if err != nil {
		return err
	}

	return fmt.Errorf("ruff format changes found:\n%s", diff)
}
