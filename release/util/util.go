package util

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

// Project is a programming language type.
type ProjectType string

const (
	Golang ProjectType = "go"
	Python ProjectType = "python"
)

// ErrUnsupportedProject indicates a project's language is not supported.
var ErrUnsupportedProject = errors.New("unsupported project type")

// ResolveProjectType determines the type of a project, returning an
// error if it's not supported.
func ResolveProjectType(language string) (ProjectType, error) {
	lang := strings.ToLower(language)

	switch lang {
	case "golang", "go":
		return Golang, nil
	case "python", "py":
		return Python, nil
	default:
		return "", fmt.Errorf("%w: recieved type %s", ErrUnsupportedProject, language)
	}
}

// ResultsFormatter provides utility for formatting sets of results.
//
//	e.g. adding a "Unit Test" header to unit test results, with a line separator before the next set of results
type ResultsFormatter interface {
	// Add appends the results content with a header.
	Add(header, content string)
	// String outputs the current state of added results.
	String() string
}

// NewResultsBasicFmt initializes a ResultsFormatter with the given separator
// used between sets of results.
func NewResultsBasicFmt(sep string) ResultsFormatter {
	return &resultsBasic{}
}

// resultsBasic is a simple implementation of ResultsFormatter.
type resultsBasic struct {
	// sep is a line separator between sets of results, e.g. '------------'.
	sep string
	// res is the running state of results, modified by Add.
	res strings.Builder
}

func (r *resultsBasic) String() string {
	return r.res.String()
}

func (r *resultsBasic) Add(header, content string) {
	r.res.Grow((len(header) + 1 + len(content) + len(r.sep) + 1))

	r.res.WriteString(header)
	r.res.WriteString("\n")
	r.res.WriteString(content)
	r.res.WriteString(r.sep)
	r.res.WriteString("\n")
}

// ExtraTags generates '<Major>, '<Major>.<Minor>', and 'latest' tags based
// on a target tag and a set of existing tags.
func ExtraTags(target string, existing []string) ([]string, error) {
	// Skip tag check if target is a prerelease
	if semver.Prerelease(target) != "" {
		return nil, nil
	}

	// check if new tag is valid semver
	if !semver.IsValid(target) {
		return nil, fmt.Errorf("new version %q is not valid semver", target)
	}

	// filter out non-semver tags and sort
	semverTags := slices.DeleteFunc(existing, func(tag string) bool {
		return !semver.IsValid(tag)
	})
	semver.Sort(semverTags)
	slices.Reverse(semverTags)

	// check if new tag doesn't already exist
	for _, tag := range semverTags {
		if tag == target {
			return nil, fmt.Errorf("version %s already exists", target)
		}
	}

	newMajor := semver.Major(target)
	newMajorMinor := semver.MajorMinor(target)

	// Find latest tags for each category.
	var latestOverall, latestMajor, latestMajorMinor bool
	for _, tag := range semverTags {
		if semver.Compare(tag, target) <= 0 {
			continue
		}
		if !latestOverall {
			latestOverall = true
		}
		if !latestMajor && semver.Major(tag) == newMajor {
			latestMajor = true
		}
		if !latestMajorMinor && semver.MajorMinor(tag) == newMajorMinor {
			latestMajorMinor = true
		}
		if latestOverall && latestMajor && latestMajorMinor {
			break
		}
	}

	publishTags := make([]string, 0, 3) // max 3

	if !latestMajorMinor {
		publishTags = append(publishTags, newMajorMinor)
	}
	if !latestMajor {
		publishTags = append(publishTags, newMajor)
	}
	if !latestOverall {
		publishTags = append(publishTags, "latest")
	}
	return publishTags, nil
}
