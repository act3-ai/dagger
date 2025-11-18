package main

import (
	"context"
	"dagger/release/internal/dagger"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Generate release notes, changelog, and VERSION file with target release version.
// Will also optionally bump a version in provided helm chart path.
func (r *Release) Prepare(ctx context.Context,
	// prepare for a specific version. Must be a valid semantic version in format of x.x.x
	version string,
	// path to helm chart in source directory to bump chart version to release version.
	// +optional
	chartPath string,
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
	//set working dir if provided
	if workingDir != "" {
		src = src.Directory(workingDir)
	}

	gcOpts := dagger.GitCliffDevOpts{
		GithubToken: githubToken,
		GitlabToken: gitlabToken,
		GiteaToken:  giteaToken,
		WorkingDir:  workingDir,
	}

	// generate changelog if one is not found, else prepend to an existing one
	changelogFile := dag.GitCliffDev(r.GitRef, gcOpts).
		Changelog(dagger.GitCliffDevChangelogOpts{
			Tag: version})

	// generate release notes with optional extraNotes if provided
	releaseNotesFile := dag.GitCliffDev(r.GitRef, gcOpts).
		ReleaseNotes(dagger.GitCliffDevReleaseNotesOpts{
			Tag:        version,
			ExtraNotes: extraNotes})

	// set helm chart version
	var chartFile *dagger.File
	if chartPath != "" {
		chartFile = r.setHelmChartVersion(version, chartPath)
	}

	// consider changing the construction of this diff
	// instead just modify the source directory directly and then compute the changes
	after := src.
		WithFile("CHANGELOG.md", changelogFile).
		WithFile(filepath.Join("releases", fmt.Sprintf("v%s.md", version)), releaseNotesFile).
		WithNewFile("VERSION", version+"\n").
		With(func(d *dagger.Directory) *dagger.Directory {
			if chartFile != nil {
				d = d.WithFile(path.Join(chartPath, "Chart.yaml"), chartFile)
			}
			return d
		})
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

	bumpedTag, err := dag.GitCliffDev(r.GitRef, dagger.GitCliffDevOpts{
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
func (r *Release) setHelmChartVersion(
	// release version
	version string,
	// path to the chart
	chartPath string,
) *dagger.File {
	version = strings.TrimPrefix(version, "v")
	file := path.Join(chartPath, "Chart.yaml")
	return dag.Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"yq"},
		}).
		WithMountedDirectory("/src", r.GitRef.Tree()).
		WithWorkdir("/src").
		WithEnvVariable("version", version).
		WithExec([]string{"yq", "e",
			"(.version = env(version)) | (.appVersion = \"v\"+env(version))",
			"-i", file}).
		File(file)
}
