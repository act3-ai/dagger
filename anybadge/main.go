// A module for generating svg badges using anybadge.
//
// This module does not require a remote service.
//
// https://pypi.org/project/anybadge/.

package main

import (
	"context"
	"dagger/anybadge/internal/dagger"
	"fmt"
	"strconv"
	"strings"
)

type Anybadge struct{}

func New() *Anybadge {
	return &Anybadge{}
}

// Generate a code coverage badge.
func (m *Anybadge) Coverage(
	// Coverage value
	value float64,
) *dagger.File {
	const (
		coverageFile = "coverage.svg"
		// color if less than:
		redThreshold    = "30"
		orangeThreshold = "50"
		yellowThreshold = "80"
		greenThreshold  = "100"
	)

	return m.Container().
		WithExec([]string{"anybadge",
			fmt.Sprintf("--value=%.2f", value),
			fmt.Sprintf("--file=%s", coverageFile),
			fmt.Sprintf("%s=red", redThreshold),
			fmt.Sprintf("%s=orange", orangeThreshold),
			fmt.Sprintf("%s=yellow", yellowThreshold),
			fmt.Sprintf("%s=green", greenThreshold),
		}).
		File(coverageFile)
}

// Generate a pylint badge.
func (m *Anybadge) Pylint(
	// Pylint value
	value float64,
) *dagger.File {
	const pylintFile = "pylint.svg"

	return m.Container().
		WithExec([]string{"anybadge",
			fmt.Sprintf("--value=%.2f", value),
			fmt.Sprintf("--file=%s", pylintFile),
			"pylint"}).
		File(pylintFile)
}

// Generate a general pipeline status badge.
func (m *Anybadge) PipelineStatus(
	// Pipeline is passing
	passing bool,
) *dagger.File {
	const daggerFile = "dagger.svg"

	value := "failing"
	if passing {
		value = "passing"
	}

	return m.Container().
		WithExec([]string{"anybadge",
			"--label=pipeline",
			fmt.Sprintf("--value=%s", value),
			fmt.Sprintf("--file=%s", daggerFile),
			"passing=green",
			"failing=red"}).
		File(daggerFile)
}

// Generate a semantic version badge.
func (m *Anybadge) Version(
	// Semantic version, e.g. "1.2.3"
	version string,
	// Badge color
	// +optional
	// +default="dodgerblue"
	color string,
) *dagger.File {
	const versionFile = "version.svg"

	// since we have the "version" label, "v" is unnecessary
	version = strings.TrimPrefix(version, "v")

	return m.Container().
		WithExec([]string{"anybadge",
			"--label=version",
			fmt.Sprintf("--value=%s", version),
			fmt.Sprintf("--file=%s", versionFile),
			fmt.Sprintf("--color=%s", color)}).
		File(versionFile)
}

// Generate a goreportcard badge. Does not rely on remote goreportcard server.
func (m *Anybadge) GoReport(ctx context.Context,
	// source code
	src *dagger.GitRef,
	// goreport reference
	// +optional
	// +defaultPath="https://github.com/gojp/goreportcard.git"
	goreportSrc *dagger.GitRef,
) *dagger.File {
	const (
		srcDir           = "src"
		goreportSrcDir   = "goreportSrc"
		goreportExecName = "goreportcard-cli"
		goreportFile     = "goreport.svg"
	)

	out, _ := dag.Go().
		Container().
		WithExec([]string{"apt", "install", "make"}).
		WithDirectory(goreportSrcDir, goreportSrc.Tree()).
		WithWorkdir(goreportSrcDir).
		// use goreport's script for installing external, vendored, deps
		WithExec([]string{"make", "install"}).
		WithExec([]string{"go", "install", "./cmd/" + goreportExecName}).
		WithDirectory("/"+srcDir, src.Tree()).
		WithWorkdir("/" + srcDir).
		WithExec([]string{goreportExecName}).
		Stdout(ctx)

	grade, value, _ := extractGradeAndPercent(out)

	var color string
	switch {
	case value >= 90:
		color = "green"
	case value >= 80:
		color = "greenyellow"
	case value >= 70:
		color = "yellow"
	case value >= 60:
		color = "orange"
	default:
		color = "red"
	}

	return m.Container().
		WithExec([]string{"anybadge",
			"--label=goreport",
			fmt.Sprintf("--value=%s", grade),
			fmt.Sprintf("--file=%s", goreportFile),
			fmt.Sprintf("--color=%s", color)}).
		File(goreportFile)
}

// Generate a license badge.
func (m *Anybadge) License(
	// License name, e.g. "MIT"
	name string,
	// Badge color
	// +optional
	// +default="darkgoldenrod"
	color string,
) *dagger.File {
	const licenseFile = "license.svg"

	return m.Container().
		WithExec([]string{"anybadge",
			"--label=license",
			fmt.Sprintf("--value=%s", name),
			fmt.Sprintf("--file=%s", licenseFile),
			fmt.Sprintf("--color=%s", color)}).
		File(licenseFile)
}

// Container returns a python container with anybadge installed.
func (m *Anybadge) Container() *dagger.Container {
	return dag.Python().
		Container().
		WithExec([]string{"pip", "install", "anybadge"})
}

func extractGradeAndPercent(report string) (grade string, percent float64, err error) {
	// should be first line
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "Grade") {
			// expected "Grade .......... A+ 100.0%"
			fields := strings.Fields(line)
			if len(fields) < 3 {
				return "", 0, fmt.Errorf("unexpected grade line format: %s", line)
			}

			grade = fields[len(fields)-2]

			percentStr := fields[len(fields)-1]
			percentStr = strings.TrimSuffix(percentStr, "%")
			percent, err = strconv.ParseFloat(percentStr, 64)
			if err != nil {
				return "", 0, fmt.Errorf("invalid percentage value: %w", err)
			}

			return grade, percent, nil
		}
	}
	return "", 0, fmt.Errorf("grade line not found")
}
