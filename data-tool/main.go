// Module for interacting with ASCE Data Tool (a.k.a., ace-dt)
package main

import (
	"context"
	"dagger/data-tool/internal/dagger"
	"fmt"
	"regexp"
	"strings"
)

const baseImage = "ghcr.io/act3-ai/data-tool:v1.16.1"

// ASCE Data Tool Module
type DataTool struct {
	Container *dagger.Container
}

func New(
	// +optional
	base *dagger.Container,
) *DataTool {
	if base != nil {
		return &DataTool{
			Container: base,
		}
	}

	// Grab the acedt executable from its image
	acedt := dag.Container().
		From(baseImage).
		// File("/ko-app/ace-dt")
		File("/usr/local/bin/ace-dt")

	const cachePath = "/oci"

	// Use the git container (instead of bash) because we want "ace-dt git" to work
	c := dag.Wolfi().Container(dagger.WolfiContainerOpts{
		Packages: []string{"bash", "git", "git-lfs"},
	}).
		WithFile("/usr/local/bin/ace-dt", acedt).
		WithUser("0").
		WithMountedCache(cachePath, dag.CacheVolume("oci-cache")).
		WithEnvVariable("ACE_DT_CACHE_PATH", cachePath)
	return &DataTool{
		Container: c,
	}
}

// Add credentials for a registry
func (m *DataTool) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *DataTool {
	regConfig := dag.RegistryConfig().WithRegistryAuth(address, username, secret)
	m.Container = m.Container.WithMountedSecret("/root/.docker/config.json", regConfig.Secret())
	return m
}

// Add credentials for use with Git via "ace-dt git"
func (m *DataTool) WithGitAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *DataTool {
	user := dag.SetSecret("username", username)
	netrc := dag.Netrc().WithLogin(address, user, secret)
	m.Container = m.Container.WithMountedSecret("/root/.netrc", netrc.AsSecret())
	return m
}

// Gathers the images and returns the full image reference with digest
func (m *DataTool) Gather(ctx context.Context,
	// artifact CSV file
	artifacts *dagger.File,

	// Destination for the gathered image as a OCI image reference
	dest string,

	// platforms
	// +optional
	platforms []dagger.Platform,
) (string, error) {
	const artifactsPath = "/artifacts.csv"

	cmd := []string{"ace-dt", "mirror", "gather", artifactsPath, dest}
	if n := len(platforms); n != 0 {
		platformStrs := make([]string, n)
		for i, platform := range platforms {
			platformStrs[i] = string(platform)
		}
		cmd = append(cmd, "--platforms", strings.Join(platformStrs, ","))
	}

	stdout, err := m.Container.
		WithFile(artifactsPath, artifacts).
		// WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec(cmd).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`Gather index: (.*)@(.*)`)

	d := re.FindAllStringSubmatch(stdout, -1)
	if len(d) != 1 {
		return "", fmt.Errorf("expected a single match for %q in: %q", re, stdout)
	}

	// return just the image reference with digest
	return fmt.Sprintf("%s@%s", d[0][1], d[0][2]), nil
}

// Scatter images to their proper locations from a gathered image
func (m *DataTool) Scatter(ctx context.Context,
	// Gathered image reference to use as the source
	ref string,

	// artifact CSV file
	mapping *dagger.File,
) error {
	// TODO support other types of mapping besides first-prefix
	// TODO support selectors
	_, err := m.Container.
		WithFile("mapping.csv", mapping).
		WithExec([]string{"ace-dt", "mirror", "scatter", ref, "first-prefix=mapping.csv"}).
		Sync(ctx)
	return err
}

// Download the Grype vulnerability database
func (m *DataTool) GrypeDB(ctx context.Context) *dagger.Directory {
	const cachePath = "/tmp/cache/grype"

	return dag.Container().
		From("anchore/grype:debug").
		// WithUser(owner).
		// WithMountedCache(cachePath, dag.CacheVolume("grype-db-cache"), dagger.ContainerWithMountedCacheOpts{Owner: owner}).
		// comment out the line below to see the cached date output
		// WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithEnvVariable("GRYPE_DB_CACHE_DIR", cachePath).
		WithExec([]string{"/grype", "db", "update"}).
		Directory(cachePath)
}

// Serialize a gathered OCI image into a TAR archive file
func (m *DataTool) Serialize(
	// OCI reference to the gathered image artifact
	ref string,
	// Include the manifest.json file (docker compatible)
	// +optional
	manifestJSON bool,
) *dagger.File {
	const archivePath = "/images.tar"

	args := []string{"ace-dt", "mirror", "serialize", ref, archivePath}

	if manifestJSON {
		args = append(args, "--manifest-json")
	}

	return m.Container.
		WithExec(args).
		File(archivePath)
}

// Archive the provided OPCI artifacts in an archive
func (m *DataTool) Archive(
	// artifact CSV file
	artifacts *dagger.File,

	// filter by platforms
	// +optional
	platforms []dagger.Platform,
	// Include the manifest.json file (docker compatible)
	// +optional
	manifestJSON bool,
) *dagger.File {
	const (
		artifactsPath = "/artifacts.csv"
		archivePath   = "/images.tar"
	)

	args := []string{"ace-dt", "mirror", "archive", artifactsPath, archivePath}

	if manifestJSON {
		args = append(args, "--manifest-json")
	}
	if len(platforms) != 0 {
		platformStrs := make([]string, len(platforms))
		for i, platform := range platforms {
			platformStrs[i] = string(platform)
		}
		args = append(args, "--platforms", strings.Join(platformStrs, ","))
	}

	return m.Container.
		WithFile(artifactsPath, artifacts).
		WithExec(args).
		File(archivePath)
}

/*
get-artifacts.sh > artifacts.csv
ace-dt mirror gather artifacts.csv registry-gitlab.dle.afrl.af.mil/cronus/arbm/sil/gathered:test
ace-dt security scan --gathered-image registry-gitlab.dle.afrl.af.mil/cronus/arbm/sil/gathered:test
*/

// TODO scan needs to depend on Gather

// Scan the images for vulnerabilities
func (m *DataTool) Scan(ctx context.Context,
	// Gathered image reference
	image string,
) (string, error) {
	grype := dag.Container().
		From("anchore/grype:latest").
		File("/grype")

	const cachePath = "/cache/grype"

	grypeDB := m.GrypeDB(ctx)

	syft := dag.Container().
		From("anchore/syft:latest").
		File("/syft")

	return m.Container.
		WithFile("/usr/local/bin/grype", grype).
		WithFile("/usr/local/bin/syft", syft).
		WithMountedDirectory(cachePath, grypeDB).
		WithEnvVariable("GRYPE_DB_CACHE_DIR", cachePath).
		// WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithUser("0").
		WithExec([]string{"ace-dt", "security", "scan", "-o=table",
			"--gathered-image", image}).
		Stdout(ctx)
}
