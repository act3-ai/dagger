// A module for creating badges using badges/shields. See https://github.com/badges/shields.
//
// Provides utilities for building common badges used in READMEs, e.g. code coverage, license, etc. The primary, default, use case is starting a shields.io image as a local dagger service, with the option to utilizie the publicly available img.shields.io.
//
// Most functions provide standard badge formats, such as 80% code coverage considered the "green" threshold.
//
// A general badge query is available if the base functions are insufficient:
//
//	dagger call send-query --label="example" --value="foo" --color="brightgreen" --remote-host= "https://img.shields.io" export --path badge.svg
package main

import (
	"context"
	"dagger/shields/internal/dagger"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Can also use docker, but we choose ghcr to avoid docker rate limits.
	// Docker image: shieldsio/shields:next
	shieldsCtr    = "ghcr.io/badges/shields:next"
	shieldsPort   = 80     // not 8080, contrary to some documentation, refer to their Dockerfile
	shieldsScheme = "http" // local service scheme
)

type Shields struct{}

func New() *Shields {
	return &Shields{}
}

// Generate a code coverage badge.
func (m *Shields) Coverage(ctx context.Context,
	// Code coverage percentage value.
	value float64,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	// https://github.com/badges/shields/blob/master/badge-maker/lib/color.js#L4
	var color string
	switch {
	case value >= 85:
		color = "brightgreen"
	case value >= 80:
		color = "green"
	case value >= 70:
		color = "yellow"
	case value >= 50:
		color = "orange"
	default:
		color = "red"
	}

	badge, _ := m.SendQuery(ctx, "coverage", "", fmt.Sprintf("%.1f", value), color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("coverage.svg")
}

// Generate a pylint badge.
func (m *Shields) Pylint(ctx context.Context,
	// Pylint score value.
	value float64,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	// https://github.com/badges/shields/blob/master/badge-maker/lib/color.js#L4
	var color string
	switch {
	case value >= 9.9:
		color = "brightgreen"
	case value >= 8.5:
		color = "green"
	case value >= 7.0:
		color = "yellow"
	case value >= 5.0:
		color = "orange"
	default:
		color = "red"
	}

	badge, _ := m.SendQuery(ctx, "pylint", "", fmt.Sprintf("%.1f", value), color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("pylint.svg")
}

// Generate a pipeline status badge.
func (m *Shields) PipelineStatus(ctx context.Context,
	// Pipeline passes.
	passing bool,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	// https://github.com/badges/shields/blob/master/badge-maker/lib/color.js#L4
	status := "failing"
	color := "red"
	if passing {
		status = "passing"
		color = "brightgreen"
	}

	badge, _ := m.SendQuery(ctx, "pipeline", "", status, color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("pipeline-status.svg")
}

// Generate a semantic version badge.
func (m *Shields) Version(ctx context.Context,
	// Badge Label
	// +optional
	// +default="version"
	label string,
	// Semantic version, e.g. "v1.2.3"
	version string,
	// Badge color
	// +optional
	// +default="blue"
	color string,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	badge, _ := m.SendQuery(ctx, label, "", version, color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("version.svg")
}

// Generate a license badge.
func (m *Shields) License(ctx context.Context,
	// License name, e.g. "MIT"
	name string,
	// Badge color. Default is a dark gold.
	// +optional
	// +default="B8860B"
	color string,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	badge, _ := m.SendQuery(ctx, "license", "", name, color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("license.svg")
}

// Generate a goreportcard badge.
func (m *Shields) GoReport(ctx context.Context,
	// source code
	src *dagger.GitRef,
	// goreport reference
	// +optional
	// +defaultPath="https://github.com/gojp/goreportcard.git"
	goreportSrc *dagger.GitRef,
	// Remote Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remoteHost, a new one is made if neither is provided.
	// +optional
	shieldsService *dagger.Service,
) *dagger.File {
	const (
		srcDir           = "src"
		goreportSrcDir   = "goreportSrc"
		goreportExecName = "goreportcard-cli"
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
		color = "brightgreen"
	case value >= 80:
		color = "green"
	case value >= 70:
		color = "yellow"
	case value >= 60:
		color = "orange"
	default:
		color = "red"
	}

	badge, _ := m.SendQuery(ctx, "goreport", "", grade, color, "", "", "", remoteHost, shieldsService)
	return badge.WithName("goreport.svg")
}

// extractGradeAndPercent parses the output of goreportcard, returning the letter
// grade and percentage value.
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

// A utility for querying a Shields service.
func (m *Shields) SendQuery(ctx context.Context,
	// Badge label.
	// +optional
	label string,
	// Badge label color.
	// +optional
	labelColor string,
	// Badge value.
	value string,
	// Badge value color. Hex, rgb, rgba, hsl, hsla and css colors.
	valueColor string,
	// Badge logo.
	// +optional
	logo string,
	// Logo color. Hex, rgb, rgba, hsl, hsla and css colors.
	// +optional
	logoColor string,
	// Badge style.
	// +optional
	style string,
	// Remotely hosted Shields service, with scheme, host, and port. Ignored if a dagger shieldsService is provided.
	// +optional
	remoteHost string,
	// (default) Shields as a dagger service. Takes precedence over remote. A new one is created if not provided and no remote specified.
	// +optional
	shieldsService *dagger.Service,
) (*dagger.File, error) {
	switch {
	case shieldsService == nil && remoteHost != "":
		// query remote
		queryURL, err := staticQuery(remoteHost, label, labelColor, value, valueColor, logo, logoColor, style)
		if err != nil {
			return nil, fmt.Errorf("building query: %w", err)
		}

		return dag.HTTP(queryURL), nil
	case shieldsService == nil:
		shieldsService = m.AsService()
		fallthrough
	default:
		const badgeFileName = "badge.svg"

		endpoint, err := shieldsService.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: shieldsPort, Scheme: shieldsScheme})
		if err != nil {
			return nil, fmt.Errorf("resolving dagger shields service endpoint: %w", err)
		}
		queryURL, err := staticQuery(endpoint, label, labelColor, value, valueColor, logo, logoColor, style)
		if err != nil {
			return nil, fmt.Errorf("building query: %w", err)
		}

		return dag.Wolfi().
			Container(dagger.WolfiContainerOpts{Packages: []string{"curl"}}).
			WithServiceBinding("shields", shieldsService).
			WithExec([]string{"curl", "-fsSL", queryURL, "-o", badgeFileName}).
			File(badgeFileName), nil
	}
}

// Shields container as a service. An optional optimization to persist the shields
// service when generating multiple badges. A new one is created if not explicitly
// provided an a remote host is not specified.
//
// Caller must use [Shields.AsService].Start and [Shields.AsService].Stop to take
// advantage of optimization.
func (m *Shields) AsService() *dagger.Service {
	return dag.Container().
		From(shieldsCtr).
		WithExposedPort(shieldsPort).
		AsService()
}

// staticQuery builds a full URL to query a shields service. Label, logo, and style are optional.
//
// Shields static badge format:
// http://<host>/badge/<label>-<message>-<color>?logo=<logo>&style=<style>
//
// See https://shields.io/badges.
func staticQuery(endpoint string, label, labelColor, value, valueColor, logo, logoColor, style string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parsing endpoint: %w", err)
	}

	// Shields static badge format:
	// /badge/<label>-<message>-<color>
	var path strings.Builder
	path.WriteString("/badge/")

	if label != "" {
		path.WriteString(formatString(label))
	}
	path.WriteString("-")
	path.WriteString(formatString(value))
	path.WriteString("-")
	path.WriteString(formatString(valueColor))

	u.Path = strings.TrimRight(u.Path, "/") + path.String()

	// Optional query parameters
	q := u.Query()
	if labelColor != "" {
		q.Set("labelColor", labelColor)
	}
	if logo != "" {
		q.Set("logo", logo)
	}
	if logoColor != "" {
		q.Set("logoColor", logo)
	}
	if style != "" {
		q.Set("style", style)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil

}

var (
	underscoreRegex = regexp.MustCompile(`_`)
	dashRegex       = regexp.MustCompile(`-`)
)

// formatString formats an partial path string as a query.
//
// See https://shields.io/badges.
func formatString(s string) string {
	// apply special rules
	s = underscoreRegex.ReplaceAllString(s, "__")
	s = dashRegex.ReplaceAllString(s, "--")
	return url.PathEscape(s)
}
