// A generated module for Shields functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

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
	// shieldsio/shields:next
	shieldsCtr    = "ghcr.io/badges/shields:next"
	shieldsPort   = 80
	shieldsScheme = "http"
)

type Shields struct {
	// shields as a service
	// +private
	svc *dagger.Service
}

func (m *Shields) Coverage(ctx context.Context,
	// Code coverage percentage value
	value float64,
) (*dagger.File, error) {
	const coverageFile = "coverage.svg"

	// https://github.com/badges/shields/blob/master/badge-maker/lib/color.js#L4
	var color string
	switch {
	case value >= 80:
		color = "green"
	case value >= 70:
		color = "yellow"
	case value >= 50:
		color = "orange"
	default:
		color = "red"
	}

	svc := m.AsService()
	svc, err := svc.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting shields as a service: %w", err)
	}
	endpoint, err := svc.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: shieldsPort, Scheme: shieldsScheme})
	if err != nil {
		return nil, fmt.Errorf("getting service endpoint: %w", err)
	}

	queryURL, err := staticQuery(endpoint, "coverage", fmt.Sprintf("%.1f", value), color, "", "")
	if err != nil {
		return nil, fmt.Errorf("building code coverage badge query: %w", err)
	}

	return dag.Wolfi().
		Container(dagger.WolfiContainerOpts{Packages: []string{"curl"}}).
		WithServiceBinding("shields", svc).
		WithExec([]string{"curl", "-fsSL", queryURL, "-o", coverageFile}).
		File(coverageFile), nil

}

// An optimization to persist the shields service when generating multiple badges.
//
// Caller must use [Shields.AsService].Start and [Shields.AsService].Stop to take
// advantage of optimization.
func (m *Shields) WithService(
	// Shields as a dagger service
	svc *dagger.Service,
) *Shields {
	m.svc = svc
	return m
}

// Shields container as a service.
func (m *Shields) AsService() *dagger.Service {
	return dag.Container().
		From(shieldsCtr).
		WithExposedPort(shieldsPort).
		AsService()
}

// staticQuery builds a full URL to query a shields service. label, logo, and style are optional.
//
// Shields static badge format:
// https://<host>/badge/<label>-<message>-<color>
func staticQuery(endpoint string, label, value, color, logo, style string) (string, error) {
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
	if style != "" {
		q.Set("style", style)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil

}

var (
	underscoreRegex = regexp.MustCompile(`_`)
	spaceRegex      = regexp.MustCompile(`\s+`)
	dashRegex       = regexp.MustCompile(`-`)
)

// formatString formats an input string as a query.
//
// See https://shields.io/badges.
func formatString(s string) string {
	// apply special rules
	s = underscoreRegex.ReplaceAllString(s, "__")
	s = dashRegex.ReplaceAllString(s, "--")
	return url.PathEscape(s)
}
