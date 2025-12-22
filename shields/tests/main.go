// A generated module for Tests functions
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
	"dagger/tests/internal/dagger"
	"fmt"
	"strings"
)

type Tests struct{}

// +check
// Run Coverage, expect no errors, validate label, value, and color.
func (t *Tests) Coverage(ctx context.Context) error {
	var tests = []struct {
		value float64
		color string
	}{
		// from inspecting svgs: https://github.com/badges/shields/tree/master/badge-maker#colors
		{90.86, "#4c1"},     // brightgreen
		{82.0, "#97ca00"},   // green
		{75.025, "#dfb317"}, // yellow
		{55.45, "#fe7d37"},  // orange
		{10.091, "#e05d44"}, // red
	}

	shieldsService := dag.Shields().AsService()
	shieldsService, err := shieldsService.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting shields service: %w", err)
	}
	defer shieldsService.Stop(ctx)

	opts := dagger.ShieldsCoverageOpts{ShieldsService: shieldsService}
	for _, t := range tests {
		svgRaw, err := dag.Shields().Coverage(t.value, opts).Contents(ctx)
		if err != nil {
			return fmt.Errorf("getting svg contents: %w", err)
		}

		// check label
		if !strings.Contains(svgRaw, ">coverage<") {
			return fmt.Errorf("expected label 'coverage' not found in svg: %s", svgRaw)
		}

		// check truncated percentage value
		expectedValue := fmt.Sprintf("%.1f", t.value)
		if !strings.Contains(svgRaw, ">"+expectedValue+"<") {
			return fmt.Errorf("expected value %q not found in svg: %s", expectedValue, svgRaw)
		}

		// check color
		if !strings.Contains(svgRaw, t.color) {
			return fmt.Errorf("expected color %q for value %f not found in svg: %s", t.color, t.value, svgRaw)
		}
	}
	return nil
}

// +check
// Run Pylint, expect no errors, validate label, value, and color.
func (t *Tests) Pylint(ctx context.Context) error {
	var tests = []struct {
		value float64
		color string
	}{
		// from inspecting svgs: https://github.com/badges/shields/tree/master/badge-maker#colors
		{9.94, "#4c1"},     // brightgreen
		{8.77, "#97ca00"},  // green
		{7.145, "#dfb317"}, // yellow
		{5.99, "#fe7d37"},  // orange
		{3.2, "#e05d44"},   // red
	}

	shieldsService := dag.Shields().AsService()
	shieldsService, err := shieldsService.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting shields service: %w", err)
	}
	defer shieldsService.Stop(ctx)

	opts := dagger.ShieldsPylintOpts{ShieldsService: shieldsService}
	for _, t := range tests {
		svgRaw, err := dag.Shields().Pylint(t.value, opts).Contents(ctx)
		if err != nil {
			return fmt.Errorf("getting svg contents: %w", err)
		}

		// check label
		if !strings.Contains(svgRaw, ">pylint<") {
			return fmt.Errorf("expected label 'coverage' not found in svg: %s", svgRaw)
		}

		// check truncated percentage value
		expectedValue := fmt.Sprintf("%.1f", t.value)
		if !strings.Contains(svgRaw, ">"+expectedValue+"<") {
			return fmt.Errorf("expected value %q not found in svg: %s", expectedValue, svgRaw)
		}

		// check color
		if !strings.Contains(svgRaw, t.color) {
			return fmt.Errorf("expected color %q for value %f not found in svg: %s", t.color, t.value, svgRaw)
		}
	}
	return nil
}

// +check
// Run PipelineStatus, expect no errors, validate label, value, and color.
func (t *Tests) PipelineStatus(ctx context.Context) error {
	var tests = []struct {
		passing bool
		color   string
	}{
		// from inspecting svgs: https://github.com/badges/shields/tree/master/badge-maker#colors
		{true, "#4c1"},     // brightgreen
		{false, "#e05d44"}, // red
	}

	shieldsService := dag.Shields().AsService()
	shieldsService, err := shieldsService.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting shields service: %w", err)
	}
	defer shieldsService.Stop(ctx)

	opts := dagger.ShieldsPipelineStatusOpts{ShieldsService: shieldsService}
	for _, t := range tests {
		svgRaw, err := dag.Shields().PipelineStatus(t.passing, opts).Contents(ctx)
		if err != nil {
			return fmt.Errorf("getting svg contents: %w", err)
		}

		// check label
		if !strings.Contains(svgRaw, ">pipeline<") {
			return fmt.Errorf("expected label 'coverage' not found in svg: %s", svgRaw)
		}

		// check value
		expectedValue := "failing"
		if t.passing {
			expectedValue = "passing"
		}
		if !strings.Contains(svgRaw, ">"+expectedValue+"<") {
			return fmt.Errorf("expected value %q not found in svg: %s", expectedValue, svgRaw)
		}

		// check color
		if !strings.Contains(svgRaw, t.color) {
			return fmt.Errorf("expected color %q for value %s not found in svg: %s", t.color, expectedValue, svgRaw)
		}
	}
	return nil
}

// +check
// Run Version, expect no errors, validate label, value, and color.
func (t *Tests) Version(ctx context.Context) error {
	const (
		defaultLabel = "version"
		defaultColor = "#007ec6" // blue
	)
	var tests = []struct {
		label   string
		version string
		color   string
	}{
		// from inspecting svgs: https://github.com/badges/shields/tree/master/badge-maker#colors
		{"", "v1.2.3", ""},        // default label ("version"), default color (blue)
		{"foo", "v1.2.3", ""},     // custom label, default color (blue)
		{"", "v1.2.3", "#4c1"},    // default label ("version"), custom color (brightgreen)
		{"bar", "v1.2.3", "#4c1"}, // custom label, custom color (brightgreen)
	}

	shieldsService := dag.Shields().AsService()
	shieldsService, err := shieldsService.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting shields service: %w", err)
	}
	defer shieldsService.Stop(ctx)

	for _, t := range tests {
		opts := dagger.ShieldsVersionOpts{
			Label:          t.label,
			Color:          strings.TrimPrefix(t.color, "#"),
			ShieldsService: shieldsService,
		}
		svgRaw, err := dag.Shields().Version(t.version, opts).Contents(ctx)
		if err != nil {
			return fmt.Errorf("getting svg contents: %w", err)
		}

		// check label
		expectedLabel := t.label
		if t.label == "" {
			expectedLabel = defaultLabel
		}
		if !strings.Contains(svgRaw, fmt.Sprintf(">%s<", expectedLabel)) {
			return fmt.Errorf("expected label %q not found in svg: %s", expectedLabel, svgRaw)
		}

		// check value
		if !strings.Contains(svgRaw, fmt.Sprintf(">%s<", t.version)) {
			return fmt.Errorf("expected value %q not found in svg: %s", t.version, svgRaw)
		}

		// check color
		expectedColor := t.color
		if t.color == "" {
			expectedColor = defaultColor
		}
		if !strings.Contains(svgRaw, expectedColor) {
			return fmt.Errorf("expected color %q not found in svg: %s", expectedColor, svgRaw)
		}
	}
	return nil
}

// +check
// Run License, expect no errors, validate label, value, and color.
func (t *Tests) License(ctx context.Context) error {
	const (
		defaultColor = "#b8860b" // dark gold
	)
	var tests = []struct {
		name  string
		color string
	}{
		// from inspecting svgs: https://github.com/badges/shields/tree/master/badge-maker#colors
		{"MIT", ""},     // default color (dark gold)
		{"foo", "#4c1"}, // custom color
	}

	shieldsService := dag.Shields().AsService()
	shieldsService, err := shieldsService.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting shields service: %w", err)
	}
	defer shieldsService.Stop(ctx)

	for _, t := range tests {
		opts := dagger.ShieldsLicenseOpts{
			Color:          strings.TrimPrefix(t.color, "#"),
			ShieldsService: shieldsService,
		}
		svgRaw, err := dag.Shields().License(t.name, opts).Contents(ctx)
		if err != nil {
			return fmt.Errorf("getting svg contents: %w", err)
		}

		// check label
		if !strings.Contains(svgRaw, ">license<") {
			return fmt.Errorf("expected label 'license' not found in svg: %s", svgRaw)
		}

		// check value
		if !strings.Contains(svgRaw, fmt.Sprintf(">%s<", t.name)) {
			return fmt.Errorf("expected value %q not found in svg: %s", t.name, svgRaw)
		}

		// check color
		expectedColor := t.color
		if t.color == "" {
			expectedColor = defaultColor
		}
		if !strings.Contains(svgRaw, expectedColor) {
			return fmt.Errorf("expected color %q not found in svg: %s", expectedColor, svgRaw)
		}
	}
	return nil
}

func (t *Tests) SendQuery(ctx context.Context) error {
	// test if we can query the public remote
	label := "foo"
	value := "bar"
	opts := dagger.ShieldsSendQueryOpts{
		Label:         label,
		RemoteService: "https://img.shields.io",
	}

	svgRaw, err := dag.Shields().SendQuery(value, value, opts).Contents(ctx)
	if err != nil {
		return fmt.Errorf("getting svg contents from public remote: %w", err)
	}

	// check label
	if !strings.Contains(svgRaw, fmt.Sprintf(">%s<", label)) {
		return fmt.Errorf("expected label %q not found in svg: %s", label, svgRaw)
	}

	// check value
	if !strings.Contains(svgRaw, fmt.Sprintf(">%s<", value)) {
		return fmt.Errorf("expected value %q not found in svg: %s", value, svgRaw)
	}

	return nil
}
