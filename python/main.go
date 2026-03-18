// A module for Python lint/testing and publishing.
// This module only supports python apps using UV for builds.
// Current linters supported: pylint, ruff, mypy, pytest, pyright

package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
	"strings"
)

const uvImageDefault = "ghcr.io/astral-sh/uv:debian"

type Python struct {
	// Base container (with cache mounts added)
	Base *dagger.Container

	// +private
	Source *dagger.Directory

	// +private
	SyncArgs []string
}

func New(
	// top-level source code directory
	// +ignore=["dist/"]
	src *dagger.Directory,
	// base development container
	// +optional
	// +defaultAddress="ghcr.io/astral-sh/uv:debian"
	base *dagger.Container,
	// extra arguments for uv sync command
	// +optional
	syncArgs []string,
) *Python {
	if syncArgs == nil {
		syncArgs = []string{
			"--frozen",
			"--all-extras",
		}
	}

	return &Python{
		Base:     base,
		Source:   src,
		SyncArgs: syncArgs,
	}
}

// base UV container with cache mounts and git-credential-helper config set
func (python *Python) base() *dagger.Container {
	return python.Base.
		WithWorkdir("/app").
		WithMountedCache("/root/.cache/uv", dag.CacheVolume("uv-cache")).
		WithMountedCache("/root/.local/share/uv", dag.CacheVolume("uv-home-cache")).
		WithEnvVariable("UV_NATIVE_TLS", "true").
		WithEnvVariable("UV_CACHE_DIR", "/root/.cache/uv"). // This is the default location for the UV_CACHE_DIR but we set it just to be safe.
		WithEnvVariable("UV_LINK_MODE", "copy").
		WithFile("/usr/local/bin/git-credential-env", dag.CurrentModule().Source().File("bin/git-credential-env.sh")). // needed for WithGitAuth()
		WithExec([]string{"git", "config", "--global", "credential.helper", "env"})                                    // needed for WithGitAuth()
}

// adds project source to base container
func (python *Python) Project() *dagger.Container {
	return python.base().WithDirectory("/app", python.Source)
}

// builds uv dependencies only from pyproject.toml and uv.lock files
func (python *Python) deps() *dagger.Container {

	return python.base().
		WithFile("/app/pyproject.toml", python.Source.File("pyproject.toml")).
		WithFile("/app/uv.lock", python.Source.File("uv.lock")).
		// WithMountedCache("/app/.venv", dag.CacheVolume("python-venv")).
		WithExec(append([]string{"uv", "sync", "--no-install-project"}, python.SyncArgs...))
}

// returns a base UV container with given source and builds dev dependencies using `uv sync`
func (python *Python) DevContainer() *dagger.Container {

	return python.deps().
		WithDirectory("/app", python.Source).
		WithExec([]string{"uv", "sync"})
}

// Add creds for private python package index
func (python *Python) WithIndexAuth(ctx context.Context,
	//name of index in pyproject.toml
	name string,
	// username to authenticate with
	username string,
	// password to authenticate with
	password *dagger.Secret,
) *Python {

	name = strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	python.Base = python.Base.WithEnvVariable(fmt.Sprintf("UV_INDEX_%s_USERNAME", name), username).
		WithSecretVariable(fmt.Sprintf("UV_INDEX_%s_PASSWORD", name), password)

	return python
}

// add an environment variable to the base container
func (python *Python) WithEnvVariable(name, value string) *Python {
	python.Base = python.Base.WithEnvVariable(name, value)
	return python
}

// adds a netrc file as a secret to the base container
func (python *Python) WithNetrc(
	//netrc file to add, in format of dagger.secret (--netrc file://mynetrc)
	netrc *dagger.Secret) *Python {
	python.Base = python.Base.WithMountedSecret("/root/.netrc", netrc)
	return python
}

// check that the lockfile is in sync with pyproject.toml
func (python *Python) CheckLock(ctx context.Context) (string, error) {
	return python.Project().
		WithExec([]string{"uv", "lock", "--check"}).
		Stdout(ctx)
}

// add credentials for private python packages from git
func (python *Python) WithGitAuth(
	// host to authenticate with e.g gitlab.com
	host string,
	// username to authenticate with
	username string,
	// password to authenticate with
	password *dagger.Secret) *Python {
	// convert host to be in proper env var format.
	host = strings.ToUpper(host)
	host = strings.ReplaceAll(host, ".", "_")
	gitUserSecret := dag.SetSecret(fmt.Sprintf("GIT_SECRET_USERNAME_%s", host), username)

	// add secret variables for provided creds
	python.Base = python.Base.WithSecretVariable(fmt.Sprintf("GIT_SECRET_USERNAME_%s", host), gitUserSecret).
		WithSecretVariable(fmt.Sprintf("GIT_SECRET_PASSWORD_%s", host), password)

	return python
}
