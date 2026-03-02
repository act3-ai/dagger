// A module for Python lint/testing and publishing.
// This module only supports python apps using UV for builds.
// Current linters supported: pylint, ruff, mypy, pytest, pyright

package main

import (
	"context"
	"dagger/python/internal/dagger"
	"fmt"
	"net/url"
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
	// +private
	gitCreds []GitCred
}

type GitCred struct {
	URL      string
	Username string
	Secret   *dagger.Secret
}

func New(
	// top-level source code directory
	// +ignore=["dist/"]
	src *dagger.Directory,
	// base development container
	// +optional
	base *dagger.Container,
	// extra arguments for uv sync command
	// +optional
	syncArgs []string,
) *Python {
	if base == nil {
		base = dag.Container().From(uvImageDefault)
	}
	// base UV container with source and cache volumes
	base = base.
		WithDirectory("/app", src).
		WithWorkdir("/app").
		WithMountedCache("/root/.cache/uv", dag.CacheVolume("uv-cache")).
		WithEnvVariable("UV_NATIVE_TLS", "true").
		WithEnvVariable("UV_CACHE_DIR", "/root/.cache/uv"). // This is the default location for the UV_CACHE_DIR but we set it just to be safe.
		WithEnvVariable("UV_LINK_MODE", "copy")

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

// returns a base UV container and builds dev dependencies using `uv sync`
func (python *Python) Container() *dagger.Container {
	return python.Base.
		WithExec(
			append(
				[]string{"uv", "sync"},
				python.SyncArgs...,
			),
		)
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
	return python.Base.
		WithExec([]string{"uv", "lock", "--check"}).
		Stdout(ctx)
}

// add credentials for private python packages from git
func (python *Python) WithGitAuth(url, username string, secret *dagger.Secret) *Python {
	python.gitCreds = append(python.gitCreds, GitCred{
		URL:      url,
		Username: username,
		Secret:   secret,
	})
	// add git credentials to the Base container
	python.Base = python.buildGitCredentialHelper(python.Base)

	return python
}

// build git-credential-env script with provided git credentials for git to use
func (python *Python) buildGitCredentialHelper(base *dagger.Container) *dagger.Container {
	if len(python.gitCreds) == 0 {
		return base
	}
	// reads Git stdin and extracts host
	baseScript := `#!/usr/bin/env sh
action="$1"
[ "$action" = "get" ] || exit 0
unset host

# Read Git input
while IFS='=' read -r key value || [ -n "$key" ]; do
	case "$key" in
		host) host="$value" ;;
	esac
done

`

	// generate cred block for each given host
	var credBlocks strings.Builder
	for i, cred := range python.gitCreds {
		u, _ := url.Parse(cred.URL)
		host := u.Host
		credBlocks.WriteString(fmt.Sprintf(`if [ "$host" = "%s" ]; then
	echo "protocol=https"
	echo "username=%s"
	echo "password=$GIT_SECRET_%d"
	exit 0
fi

`, host, cred.Username, i))
	}

	fullScript := baseScript + credBlocks.String() + "echo\n"

	// add credential script to container
	base = base.WithNewFile("/usr/local/bin/git-credential-env", fullScript,
		dagger.ContainerWithNewFileOpts{Permissions: 0755})

	// add secret variables for provided gitCreds
	for i, cred := range python.gitCreds {
		base = base.WithSecretVariable(fmt.Sprintf("GIT_SECRET_%d", i), cred.Secret)
	}

	// configure git to use credential helper script
	base = base.WithExec([]string{"git", "config", "--global", "credential.helper", "env"})

	return base
}
