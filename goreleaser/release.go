package main

import (
	"context"
	"dagger/goreleaser/internal/dagger"
	"errors"
	"fmt"
	"strconv"
)

// Release represents the `goreleaser release` command.
type Release struct {
	// +private
	Goreleaser *Goreleaser

	// release command
	// +private
	Command []string
}

// Release represents the `goreleaser release` command.
func (gr *Goreleaser) Release() *Release {
	return &Release{
		Goreleaser: gr,
		Command:    []string{"goreleaser", "release"},
	}
}

// Run `goreleaser release` with all options previously provided.
//
// Run MAY be used as a "catch-all" in case functions are not implemented.
func (r *Release) Run(ctx context.Context,
	// arguments and flags, without `git-cliff`
	// +optional
	args []string,
	// Output results, without an error.
	// +optional
	ignoreError bool,
) (string, error) {
	// We could validate the config within New(), failing slightly earlier, but
	// running dagger with '--silent' returns a vague error and a panic is too harsh.
	// So we choose here so we can be a bit more informative.
	if err := r.Goreleaser.checkConfig(ctx); err != nil {
		return "", err
	}

	cmd := r.Command
	cmd = append(cmd, args...)
	out, err := r.Goreleaser.Container.
		WithExec(cmd).
		Stdout(ctx)

	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		// exit code != 0
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		if ignoreError {
			return result, nil
		}
		return "", fmt.Errorf("%s", result)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// exit code 0
		return out, nil
	}
}

// Generate an unversioned snapshot release, skipping all validations and without publishing any artifacts.
//
// e.g. `goreleaser release --snapshot`.
func (r *Release) WithSnapshot() *Release {
	r.Command = append(r.Command, "--snapshot")
	return r
}

// Automatically sets WithSnapshot if the repository is dirty.
//
// e.g. `goreleaser build --auto-snapshot`.
func (r *Release) WithAutoSnapshot() *Release {
	r.Command = append(r.Command, "--auto-snapshot")
	return r
}

// Removes the 'dist' directory before building.
//
// e.g. `goreleaser release --clean`.
func (r *Release) WithClean() *Release {
	r.Command = append(r.Command, "--clean")
	return r
}

// Timeout to the entire release process.
//
// e.g. `goreleaser build --timeout <duration>`.
func (r *Release) WithTimeout(
	// Timeout duration, e.g. 10m, 10m30s. Default is 30m.
	duration string,
) *Release {
	r.Command = append(r.Command, "--timeout", duration)
	return r
}

// Abort the release publishing on the first error.
//
// e.g. `goreleaser release --fail-fast`.
func (r *Release) WithFailFast() *Release {
	r.Command = append(r.Command, "--fail-fast")
	return r
}

// Tasks to run concurrently (default: number of CPUs).
//
// e.g. `goreleaser release --parallelism <n>`.
func (r *Release) WithParallelism(
	// concurrent tasks
	n int,
) *Release {
	r.Command = append(r.Command, "--parallelism", strconv.Itoa(n))
	return r
}

// Load custom release notes from a markdown file, skips changelong generation.
//
// e.g. `goreleaser release --release-notes <notes>`.
func (r *Release) WithNotes(
	// release notes markdown file
	notes *dagger.File,
) *Release {
	notesPath := "/work/notes.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(notesPath, notes)
	r.Command = append(r.Command, "--release-notes", notesPath)
	return r
}

// Load custom release notes from a templated markdown file. Overrides WithNotes.
//
// e.g. `goreleaser release --release-notes-tmpl <notesTmpl>`.
func (r *Release) WithNotesTmpl(
	// release notes templated markdown file
	notesTmpl *dagger.File,
) *Release {
	notesPath := "/work/notes-tmpl.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(notesPath, notesTmpl)
	r.Command = append(r.Command, "--release-notes-tmpl", notesPath)
	return r
}

// Load custom release notes header from a markdown file.
//
// e.g. `goreleaser release --release-header <header>`.
func (r *Release) WithNotesHeader(header *dagger.File) *Release {
	headerPath := "/work/header.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(headerPath, header)
	r.Command = append(r.Command, "--release-header", headerPath)
	return r
}

// Load custom release notes header from a templated markdown file. Overrides WithNotesHeader.
//
// e.g. `goreleaser release --release-header-tmpl <headerTmpl>`.
func (r *Release) WithNotesHeaderTmpl(
	// release notes header templated markdown file
	headerTmpl *dagger.File,
) *Release {
	headerPath := "/work/header-tmpl.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(headerPath, headerTmpl)
	r.Command = append(r.Command, "release-header-tmpl", headerPath)
	return r
}

// Load custom release notes footer from a markdown file.
//
// e.g. `goreleaser release --release-footer <footer>`.
func (r *Release) WithNotesFooter(footer *dagger.File) *Release {
	footerPath := "/work/header.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(footerPath, footer)
	r.Command = append(r.Command, "--release-footer", footerPath)
	return r
}

// Load custom release notes footer from a templated markdown file. Overrides WithNotesFooter.
//
// e.g. `goreleaser release --release-footer-tmpl <footerTmpl>`.
func (r *Release) WithNotesFooterTmpl(
	// release notes footer templated markdown file
	footerTmpl *dagger.File,
) *Release {
	footerPath := "/work/footer-tmpl.md"
	r.Goreleaser.Container = r.Goreleaser.Container.WithMountedFile(footerPath, footerTmpl)
	r.Command = append(r.Command, "--release-footer-tmpl", footerPath)
	return r
}
