// A module for generating svg badges using anybadge.
//
// https://pypi.org/project/anybadge/.

package main

import (
	"context"
	"dagger/anybadge/internal/dagger"
	"fmt"
)

const badgeFile = "badge.svg"

type Anybadge struct{}

func New() *Anybadge {
	return &Anybadge{}
}

// Run runs anybadge, returning the badge svg.
func (m *Anybadge) Run(ctx context.Context,
	// Badge label.
	// +optional
	label string,
	// Badge value.
	// +optional
	value string,
	// Value formatting. Ex: "%.2f" for 2dp floats.
	// +optional
	valueFormat string,
	// Fixed color, disregards thresholds.
	// +optional
	color string,
	// Value prefix.
	// +optional
	prefix string,
	// Value suffix.
	// +optional
	suffix string,
	// Number of chars to pad on either size of text.
	// +optional
	padding string,
	// Number of chars to pad on either side of label.
	// +optional
	labelPadding string,
	// Number of chars to pad on either side of value.
	// +optional
	valuePadding string,
	// Text font. Supports: Arial, Helvetica, DejaVu Sans, Verdana, Geneva, ans-serif.
	// +optional
	font string,
	// Text font size.
	// +optional
	fontSize int,
	// Built in template name, e.g. pylint, coverage.
	// +optional
	template string,
	// Alternative badge style. Supports: gitlab-scoped, default.
	// +optional
	style string,
	// Use maximum threshold color when value exceeds threshold.
	// +optional
	useMax bool,
	// Text color. Single value affects both label and value, a comma separated pair affects label and value respectively.
	// +optional
	textColor string,
	// Treat value and thresholds as semantic versions.
	// +optional
	semver bool,
	// Do not escape the label text.
	// +optional
	noEscapedLabel bool,
	// Do not escape the value text.
	// +optional
	noEscapedValue bool,
	// Threshold args, pairs of <upper>=<color>. Ex: 2=red 4=orange 6=yellow 8=good, read as "less than 2 = red, less than 4 = orange, ...". https://github.com/jongracecox/anybadge?tab=readme-ov-file#colors.
	// +optional
	thresholds []string,
) *dagger.File {
	ctr := dag.Python().
		Container().
		WithExec([]string{"pip", "install", "anybadge"})

	args := make([]string, 0, len(thresholds)+3) // at a minimum we have "anybadge" + "--file" + thresholds + (label or value)
	args = append(args, "anybadge")
	args = append(args, fmt.Sprintf("--file=%s", badgeFile))

	if label != "" {
		args = append(args, fmt.Sprintf("--label=%s", label))
	}

	if value != "" {
		args = append(args, fmt.Sprintf("--value=%s", value))
	}

	if valueFormat != "" {
		args = append(args, fmt.Sprintf("--value-format=%s", valueFormat))
	}

	if color != "" {
		args = append(args, fmt.Sprintf("--color=%s", color))
	}

	if prefix != "" {
		args = append(args, fmt.Sprintf("--prefix=%s", prefix))
	}

	if suffix != "" {
		args = append(args, fmt.Sprintf("--suffix=%s", suffix))
	}

	if padding != "" {
		args = append(args, fmt.Sprintf("--padding=%s", padding))
	}

	if labelPadding != "" {
		args = append(args, fmt.Sprintf("--label-padding=%s", labelPadding))
	}

	if valuePadding != "" {
		args = append(args, fmt.Sprintf("--value-padding=%s", valuePadding))
	}

	if font != "" {
		args = append(args, fmt.Sprintf("--font=%s", font))
	}

	if fontSize > 0 {
		args = append(args, fmt.Sprintf("--font-size=%d", fontSize))
	}

	if template != "" {
		args = append(args, fmt.Sprintf("--template=%s", template))
	}

	if style != "" {
		args = append(args, fmt.Sprintf("--style=%s", style))
	}

	if useMax {
		args = append(args, "--use-max")
	}

	if textColor != "" {
		args = append(args, fmt.Sprintf("--text-color=%s", textColor))
	}

	if semver {
		args = append(args, "--semver")
	}

	if noEscapedLabel {
		args = append(args, "--no-escape-label")
	}

	if noEscapedValue {
		args = append(args, "--no-escape-value")
	}

	if len(thresholds) > 0 {
		args = append(args, thresholds...)
	}

	return ctr.WithExec(args).File(badgeFile)
}
