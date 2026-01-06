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

// Runs markdownlint-cli2 against a given source directory. Returns a container that with fail with any errors.
func (m *Markdownlint) Lint(
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,
) *dagger.Container {
	cmd := m.Command
	cmd = append(cmd, extraArgs...)

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}

	ctr := m.Base.WithExec(cmd)

	return ctr

}

// Runs markdownlint-cli2 and returns results in a file.
func (m *Markdownlint) Report(
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string,
) *dagger.File {
	cmd := m.Command
	cmd = append(cmd, extraArgs...)

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}

	results := m.Base.WithExec(cmd,
		dagger.ContainerWithExecOpts{
			Expect:         dagger.ReturnTypeAny,
			RedirectStdout: "markdownlint-results.txt"}).
		File("markdownlint-results.txt")

	return results

}

// AutoFix attempts to fix any linting errors reported by rules that emit fix information.
// Returns a Changeset that can be used to apply any changes made
// to the host.
// e.g. 'markdownlint-cli2 --fix'.
func (m *Markdownlint) AutoFix(
	// Additional arguments to pass to markdownlint-cli2, without 'markdownlint-cli2' itself.
	// +optional
	extraArgs []string) *dagger.Changeset {
	cmd := m.Command
	cmd = append(cmd, "--fix")

	if !m.disableDefaultGlobs || len(extraArgs) <= 0 {
		// match all markdown files, see "Dot-only glob" https://github.com/DavidAnson/markdownlint-cli2?tab=readme-ov-file#command-line
		cmd = append(cmd, ".")
	}
	ctr := m.Base.WithUser("root").
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny})

	afterChanges := ctr.Directory("/work/src").Filter(dagger.DirectoryFilterOpts{Exclude: []string{""}})

	return afterChanges.Changes(m.Base.Directory("/work/src"))
}
