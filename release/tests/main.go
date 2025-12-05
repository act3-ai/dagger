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

	"github.com/dagger/dagger/util/parallel"
)

type Tests struct{}

// Run all tests.
func (t *Tests) All(ctx context.Context) error {
	return parallel.New().
		WithJob("Prepare", t.Prepare).
		WithJob("Prepare helm chart", t.PrepareHelmChart).
		Run(ctx)
}

// return container with a git repo and an initial commit with tag v1.0.0
func (t *Tests) gitRepo() *dagger.Container {

	return dag.Container().
		From("alpine/git").
		WithWorkdir("/repo").
		WithNewFile("README.md", "# My Project").
		WithExec([]string{"git", "init"}).
		WithExec([]string{"git", "config", "user.name", "test"}).
		WithExec([]string{"git", "config", "user.email", "test@dagger.io"}).
		WithFile("cliff.toml", dag.CurrentModule().Source().File("test.cliff.toml")).
		WithExec([]string{"git", "add", "README.md", "cliff.toml"}).
		WithExec([]string{"git", "commit", "-m", "fix: Initial commit"}).
		WithExec([]string{"git", "tag", "-a", "-m", "Initial commit", "v1.0.0"})

}

// ensure prepare generates a CHANGELOG.md, VERSION, and releases/v1.0.1.md file after a bump
func (t *Tests) Prepare(ctx context.Context) error {

	gitref := t.gitRepo().
		WithNewFile("test.md", "test").
		WithExec([]string{"git", "add", "test.md"}).
		WithExec([]string{"git", "commit", "-m", "fix: test tag"}).Directory("/repo").AsGit().Head()

	const expectedPatch = `diff --git b/CHANGELOG.md b/CHANGELOG.md
new file mode 100644
index 0000000..9427169
--- /dev/null
+++ b/CHANGELOG.md
@@ -0,0 +1,5 @@
+## [1.0.1]
+
+### 🐛 Bug Fixes
+
+- Test tag
diff --git b/VERSION b/VERSION
new file mode 100644
index 0000000..7dea76e
--- /dev/null
+++ b/VERSION
@@ -0,0 +1 @@
+1.0.1
diff --git b/releases/v1.0.1.md b/releases/v1.0.1.md
new file mode 100644
index 0000000..9427169
--- /dev/null
+++ b/releases/v1.0.1.md
@@ -0,0 +1,5 @@
+## [1.0.1]
+
+### 🐛 Bug Fixes
+
+- Test tag
`
	version, err := dag.Release(gitref).Version(ctx)
	changes := dag.Release(gitref).Prepare(version)

	patch, err := changes.AsPatch().Contents(ctx)

	if expectedPatch != patch {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", patch, expectedPatch)
	}

	return err
}

func (t *Tests) PrepareHelmChart(ctx context.Context) error {
	const chartYaml = `apiVersion: v2
name: mychart
description: A Helm chart for my cool stuff.
type: application
version: 1.4.1
appVersion: "v1.4.1"
`

	gitref := t.gitRepo().
		WithNewFile("charts/mychart/Chart.yaml", chartYaml).
		WithExec([]string{"git", "add", "charts"}).
		WithExec([]string{"git", "commit", "-m", "fix: test commit"}).Directory("/repo").AsGit().Head()

	const expectedPatch = `diff --git a/charts/mychart/Chart.yaml b/charts/mychart/Chart.yaml
index e539fc7..2ac64b8 100644
--- a/charts/mychart/Chart.yaml
+++ b/charts/mychart/Chart.yaml
@@ -2,5 +2,5 @@ apiVersion: v2
 name: mychart
 description: A Helm chart for my cool stuff.
 type: application
-version: 1.4.1
-appVersion: "v1.4.1"
+version: 1.5.6
+appVersion: "v1.5.6"
`

	changes := dag.Release(gitref).PrepareHelmChart("v1.5.6", "charts/mychart")

	patch, err := changes.AsPatch().Contents(ctx)

	if expectedPatch != patch {
		return fmt.Errorf("unexpected patch\nACTUAL:\n%s\nEXPECTED:\n%s\n", patch, expectedPatch)
	}

	return err
}
