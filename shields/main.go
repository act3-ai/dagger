// A module for creating badges using badges/sheilds.
//
// Provides utilities for building common badges used in READMEs, e.g. code coverage, license, etc. Capable of using a Sheilds image as a dagger service or the publicly available img.shields.io.
//
// Public img.shields.io example:
//
//	dagger call send-query --label="example" --value="foo" --color="brightgreen" --remote-service="https://img.shields.io" export --path badge.svg
//
// See https://github.com/badges/shields.
package main

import (
	"context"
	"dagger/shields/internal/dagger"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	// Can also use docker, but we choose ghcr to avoid docker rate limits.
	// Docker image: shieldsio/shields:next
	shieldsCtr    = "ghcr.io/badges/shields:next"
	shieldsPort   = 80 // not 8080, contrary to some documentation, refer to their Dockerfile
	shieldsScheme = "http"
)

type Shields struct{}

// Generate a code coverage badge.
func (m *Shields) Coverage(ctx context.Context,
	// Code coverage percentage value.
	value float64,
	// Remote Sheilds service, with scheme, host, and port. Ignored if a dagger sheildsService is provided.
	// +optional
	remoteService string,
	// Sheilds as a dagger service, a new one is made if not provided. An optimization.
	// +optional
	sheildsService *dagger.Service,
) *dagger.File {
	const coverageFile = "coverage.svg"

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

	badge, _ := m.SendQuery(ctx, "coverage", fmt.Sprintf("%.1f", value), color, "", "", "", remoteService, sheildsService)
	return badge
}

// Generate a pylint badge.
func (m *Shields) Pylint(ctx context.Context,
	// Pylint score value.
	value float64,
	// Remote Sheilds service, with scheme, host, and port. Ignored if a dagger sheildsService is provided.
	// +optional
	remoteService string,
	// Sheilds as a dagger service, a new one is made if not provided. An optimization.
	// +optional
	sheildsService *dagger.Service,
) *dagger.File {
	const coverageFile = "pylint.svg"

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

	badge, _ := m.SendQuery(ctx, "pylint", fmt.Sprintf("%.1f", value), color, "", "", "", remoteService, sheildsService)
	return badge
}

// Generate a pipeline status badge.
func (m *Shields) PipelineStatus(ctx context.Context,
	// Pipeline passes.
	passing bool,
	// Remote Sheilds service, with scheme, host, and port. Ignored if a dagger sheildsService is provided.
	// +optional
	remoteService string,
	// Sheilds as a dagger service, a new one is made if not provided. An optimization.
	// +optional
	sheildsService *dagger.Service,
) *dagger.File {
	const coverageFile = "pipeline.svg"

	// https://github.com/badges/shields/blob/master/badge-maker/lib/color.js#L4
	status := "failing"
	color := "red"
	if passing {
		status = "passing"
		color = "brightgreen"
	}

	badge, _ := m.SendQuery(ctx, "pipeline", status, color, "", "", "", remoteService, sheildsService)
	return badge
}

// Generate a semantic version badge.
func (m *Shields) Version(ctx context.Context,
	// Semantic version, e.g. "1.2.3", cleans "v" prefix.
	version string,
	// Badge color
	// +optional
	// +default="blue"
	color string,
	// Remote Sheilds service, with scheme, host, and port. Ignored if a dagger sheildsService is provided.
	// +optional
	remoteService string,
	// Sheilds as a dagger service, a new one is made if not provided. An optimization.
	// +optional
	sheildsService *dagger.Service,
) *dagger.File {
	const coverageFile = "version.svg"

	version = strings.TrimPrefix(version, "v")

	badge, _ := m.SendQuery(ctx, "version", version, color, "", "", "", remoteService, sheildsService)
	return badge
}

// A utility for querying a Sheilds service.
func (m *Shields) SendQuery(ctx context.Context,
	// Badge label.
	// +optional
	label string,
	// Badge value.
	value string,
	// Badge color. Hex, rgb, rgba, hsl, hsla and css colors.
	color string,
	// Badge logo.
	// +optional
	logo string,
	// Logo color. Hex, rgb, rgba, hsl, hsla and css colors.
	// +optional
	logoColor string,
	// Badge style.
	// +optional
	style string,
	// Remote Sheilds service, with scheme, host, and port. Ignored if a dagger sheildsService is provided.
	// +optional
	remoteService string,
	// Sheilds as a dagger service. Takes precedence over remote. A new one is created if not provided and no remote specified.
	// +optional
	sheildsService *dagger.Service,
) (*dagger.File, error) {
	switch {
	case sheildsService == nil && remoteService != "":
		// query remote
		queryURL, err := staticQuery(remoteService, label, value, color, logo, logoColor, style)
		if err != nil {
			return nil, fmt.Errorf("building query: %w", err)
		}

		return dag.HTTP(queryURL), nil
	case sheildsService == nil:
		sheildsService = m.AsService()
		fallthrough
	default:
		const badgeFileName = "badge.svg"

		endpoint, err := sheildsService.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: shieldsPort, Scheme: shieldsScheme})
		if err != nil {
			return nil, fmt.Errorf("resolving dagger sheilds service endpoint: %w", err)
		}
		queryURL, err := staticQuery(endpoint, label, value, color, logo, logoColor, style)
		if err != nil {
			return nil, fmt.Errorf("building query: %w", err)
		}

		return dag.Wolfi().
			Container(dagger.WolfiContainerOpts{Packages: []string{"curl"}}).
			WithServiceBinding("shields", sheildsService).
			WithExec([]string{"curl", "-fsSL", queryURL, "-o", badgeFileName}).
			File(badgeFileName), nil
	}
}

// Shields container as a service. An optimization to persist the shields service when generating multiple badges.
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
func staticQuery(endpoint string, label, value, color, logo, logoColor, style string) (string, error) {
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
	path.WriteString(formatString(color))

	u.Path = strings.TrimRight(u.Path, "/") + path.String()

	// Optional query parameters
	q := u.Query()
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
