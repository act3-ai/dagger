package main

import (
	"context"
	"dagger/release/internal/dagger"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Generate release notes, changelog, and VERSION file with target release version.
// Will also optionally bump a version in provided helm chart path.
func (r *Release) Prepare(ctx context.Context,
	// prepare for a specific version. Must be a valid semantic version in format of x.x.x
	version string,
	// Additional information to include in release notes. Injected after header and before commit
	// +optional
	extraNotes string,
	//Working Directory in source directory to run git-cliff
	// +optional
	workingDir string,
	//provide a github token to git-cliff.
	// +optional
	githubToken *dagger.Secret,
	//provide a gitlab token to git-cliff.
	// +optional
	gitlabToken *dagger.Secret,
	//provide a gitea token to git-cliff.
	// +optional
	giteaToken *dagger.Secret,
) (*dagger.Changeset, error) {
	src := r.GitRef.Tree(dagger.GitRefTreeOpts{Depth: -1})

	version = strings.TrimPrefix(version, "v")

	//check if provided version is valid
	_, err := semver.StrictNewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid semver %q: %w", version, err)
	}

	gcOpts := dagger.GitCliffOpts{
		GithubToken: githubToken,
		GitlabToken: gitlabToken,
		GiteaToken:  giteaToken,
		WorkingDir:  workingDir,
	}

	// generate changelog if one is not found, else prepend to an existing one
	changelogFile := dag.GitCliff(r.GitRef, gcOpts).
		Changelog(dagger.GitCliffChangelogOpts{
			Tag: version})

	// generate release notes with optional extraNotes if provided
	releaseNotesFile := dag.GitCliff(r.GitRef, gcOpts).
		ReleaseNotes(dagger.GitCliffReleaseNotesOpts{
			Tag:        version,
			ExtraNotes: extraNotes})

	// consider changing the construction of this diff
	// instead just modify the source directory directly and then compute the changes
	after := src.
		WithFile(path.Join(workingDir, "CHANGELOG.md"), changelogFile).
		WithFile(path.Join(workingDir, "releases", fmt.Sprintf("v%s.md", version)), releaseNotesFile).
		WithNewFile(path.Join(workingDir, "VERSION"), version+"\n")

	return after.Changes(src), nil
}

// regex to confirm valid semver
var semverRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9_-]+/)?v?(\d+\.\d+\.\d+(?:-[0-9A-Za-z-.]+)?)`)

// Generate the next version from conventional commit messages using git-cliff.
// Will attempt to coerce a bumped tag if not in semantic version format and
// Returns a version in format of MAJOR.MINOR.PATCH ex: 1.0.0
func (r *Release) Version(ctx context.Context,
	//Working Directory in source directory to run git-cliff
	// +optional
	workingDir string,
	//provide a github token to git-cliff.
	// +optional
	githubToken *dagger.Secret,
	//provide a gitlab token to git-cliff.
	// +optional
	gitlabToken *dagger.Secret,
	//provide a gitea token to git-cliff.
	// +optional
	giteaToken *dagger.Secret,
) (string, error) {

	bumpedTag, err := dag.GitCliff(r.GitRef, dagger.GitCliffOpts{
		GithubToken: githubToken,
		GitlabToken: gitlabToken,
		GiteaToken:  giteaToken,
		WorkingDir:  workingDir}).
		BumpedVersion(ctx)

	if bumpedTag == "" {
		return "", fmt.Errorf("there was nothing to bump")
	}

	semver := semverRegex.FindStringSubmatch(bumpedTag)

	if len(semver) < 2 {
		return "", fmt.Errorf("valid semver not found in: %s", bumpedTag)
	}
	return semver[1], err

}

// Set the version and appVersion of a helm chart.
func (r *Release) SetHelmChartVersion(
	// release version
	version string,
	// path to the chart
	chartPath string,
) *dagger.Changeset {
	src := r.GitRef.Tree(dagger.GitRefTreeOpts{Depth: -1})
	version = strings.TrimPrefix(version, "v")
	file := path.Join(chartPath, "Chart.yaml")
	chart := dag.Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"yq"},
		}).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("version", version).
		WithExec([]string{"yq", "e",
			"(.version = env(version)) | (.appVersion = \"v\"+env(version))",
			"-i", file}).
		File(file)

	return src.WithFile(file, chart).Changes(src)
}
