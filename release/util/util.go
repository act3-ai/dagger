package util

import (
	"errors"
	"fmt"
	"slices"

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
