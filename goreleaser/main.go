// A module for running the Goreleaser CLI.
//
// This module aids in building executables and publishing releases to public
// or private remotes. The bulk of configuration should be done in a .goreleaser.yaml
// file.

package main

import (
	"context"
	"dagger/goreleaser/internal/dagger"
	"errors"
	"fmt"
	"os"
)

// environment variable names
const (
	envGOMAXPROCS = "GOMAXPROCS"
	envGOMEMLIMIT = "GOMEMLIMIT"
	envGOOS       = "GOOS"
	envGOARCH     = "GOARCH"
	envGOARM      = "GOARM"
)

const (
	imageGoReleaser = "ghcr.io/goreleaser/goreleaser" // defaults to "latest"
)

// Goreleaser represents the `goreleaser` command.
type Goreleaser struct {
	Container *dagger.Container

	// +private
	RegistryConfig *dagger.RegistryConfig
}

func New(ctx context.Context,
	// Git repository source.
	src *dagger.Directory,

	// Additonal .gitignore file
	// +optional
	gitIgnore *dagger.File,

	// Custom container to use as a base container. Must have 'goreleaser' available on PATH.
	// +optional
	container *dagger.Container,

	// Version (image tag) to use as a goreleaser binary source.
	// +optional
	// +default="latest"
	version string,

	// Configuration file.
	// +optional
	config *dagger.File,

	// Mount netrc credentials for a private git repository.
	// +optional
	netrc *dagger.Secret,

	// Disable mounting cache volumes.
	//
	// +optional
	disableCache bool,
) *Goreleaser {
	if container == nil {
		container = defaultContainer(version)
	}

	flags := []string{"goreleaser"}
	srcDir := "/work/src"
	container = container.With(
		func(c *dagger.Container) *dagger.Container {
			if config != nil {
				cfgPath, err := config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, config)
				flags = append(flags, "--config", cfgPath)
			}
			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if !disableCache {
				c = withGoModuleCacheFn(dag.CacheVolume("go-mod"), nil, "")(c)
				c = withGoBuildCacheFn(dag.CacheVolume("go-build"), nil, "")(c)
			}
			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if netrc != nil {
				c = c.WithMountedSecret("/root/.netrc", netrc)
			}
			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if gitIgnore != nil {
				const gitIgnorePath = "/work/.gitignore"
				c = c.WithMountedFile(gitIgnorePath, gitIgnore).
					WithExec([]string{"git", "config", "--global", "core.excludesfile", gitIgnorePath})
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, src)

	gr := &Goreleaser{
		Container:      container,
		RegistryConfig: dag.RegistryConfig(),
	}

	return gr
}

// WithEnvVariable adds an environment variable to the goreleaser container.
//
// This is useful for reusability and readability by not breaking the goreleaser calling chain.
func (gr *Goreleaser) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
	// The value of the environment variable (e.g., "localhost").
	value string,
	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) *Goreleaser {
	gr.Container = gr.Container.WithEnvVariable(
		name,
		value,
		dagger.ContainerWithEnvVariableOpts{
			Expand: expand,
		},
	)
	return gr
}

// WithSecretVariable adds an env variable containing a secret to the goreleaser container.
//
// This is useful for reusability and readability by not breaking the goreleaser calling chain.
func (gr *Goreleaser) WithSecretVariable(
	// The name of the environment variable containing a secret (e.g., "PASSWORD").
	name string,
	// The value of the environment variable containing a secret.
	secret *dagger.Secret,
) *Goreleaser {
	gr.Container = gr.Container.WithSecretVariable(name, secret)
	return gr
}

// Add registry credentials.
func (gr *Goreleaser) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *Goreleaser {
	gr.RegistryConfig = gr.RegistryConfig.WithRegistryAuth(address, username, secret)
	return gr
}

// Mount a cache volume for Go module cache.
func (gr *Goreleaser) WithGoModuleCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *Goreleaser {
	gr.Container = withGoModuleCacheFn(cache, source, sharing)(gr.Container)
	return gr
}

// Mount a cache volume for Go build cache.
func (gr *Goreleaser) WithGoBuildCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *Goreleaser {
	gr.Container = withGoBuildCacheFn(cache, source, sharing)(gr.Container)
	return gr
}

// Run goreleaser.
//
// Run is a "catch-all" in case functions are not implemented.
func (gr *Goreleaser) Run(ctx context.Context,
	// arguments and flags, without `goreleaser`.
	args []string,
	// Output results, without an error.
	// +optional
	ignoreError bool,
) (string, error) {
	// We could validate the config within New(), failing slightly earlier, but
	// running dagger with '--silent' returns a vague error and a panic is too harsh.
	// So we choose here so we can be a bit more informative.
	if err := gr.checkConfig(ctx); err != nil {
		return "", err
	}

	out, err := gr.Container.WithExec(append([]string{"goreleaser"}, args...)).Stdout(ctx)
	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		// exit code != 0
		result := fmt.Sprintf("Stout:\n%s\n\nStderr:\n%s", e.Stdout, e.Stderr)
		if ignoreError {
			return result, nil
		}
		return "", fmt.Errorf("%s", result)
	case err != nil:
		// some other dagger error, e.g. graphql
		return "", err
	default:
		// exit code 0
		return out, nil
	}
}

// defaultContainer constructs a minimal container containing a source git repository.
func defaultContainer(version string) *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("%s:%s", imageGoReleaser, version)).
		With(func(r *dagger.Container) *dagger.Container {
			// inherit from host, overriden by WithEnvVariable
			val, ok := os.LookupEnv(envGOMAXPROCS)
			if ok {
				r = r.WithEnvVariable(envGOMAXPROCS, val)
			}
			return r
		}).
		With(func(r *dagger.Container) *dagger.Container {
			// inherit from host, overriden by WithEnvVariable
			val, ok := os.LookupEnv(envGOMEMLIMIT)
			if ok {
				r = r.WithEnvVariable(envGOMEMLIMIT, val)
			}
			return r
		})
}

// withGoModuleCacheFn is a helper func, allowing us to use it directly on containrs and easily expose it with 'WithGoModuleCache'.
func withGoModuleCacheFn(
	cache *dagger.CacheVolume,
	source *dagger.Directory,
	sharing dagger.CacheSharingMode,
) func(c *dagger.Container) *dagger.Container {
	return func(c *dagger.Container) *dagger.Container {
		c = c.WithMountedCache(
			"/go/pkg/mod",
			cache,
			dagger.ContainerWithMountedCacheOpts{
				Source:  source,
				Sharing: sharing,
			},
		)

		return c
	}
}

// withGoBuildCacheFn is a helper func, allowing us to use it directly on containrs and easily expose it with 'WithGoBuildCache'.
func withGoBuildCacheFn(
	cache *dagger.CacheVolume,
	source *dagger.Directory,
	sharing dagger.CacheSharingMode,
) func(c *dagger.Container) *dagger.Container {
	return func(c *dagger.Container) *dagger.Container {
		c = c.WithMountedCache(
			"/root/.cache/go-build",
			cache,
			dagger.ContainerWithMountedCacheOpts{
				Source:  source,
				Sharing: sharing,
			},
		)

		return c
	}
}

// checkConfig validates a goreleaser config. This is mostly to keep the user sane.
// Given an invalid config goreleaser will throw an error for the problematic line
// but reports it as an error for the command being run, e.g. a build error.
func (gr *Goreleaser) checkConfig(ctx context.Context) error {
	// results of check always on stderr
	_, err := gr.Container.WithExec([]string{"goreleaser", "check"}).Stdout(ctx)
	// "parse" the error for useful info
	var e *dagger.ExecError
	switch {
	case errors.As(err, &e):
		return fmt.Errorf("%s", e.Stderr)
	case err != nil:
		// some other dagger error, e.g. graphql
		return err
	default:
		// exit code 0, no issues found
		return nil
	}
}
