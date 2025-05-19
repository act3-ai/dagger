// A module for running the Goreleaser CLI.
//
// This module aids in building executables and releasing. The bulk of configuration
// should be done in a .goreleaser.yaml file.

package main

import (
	"context"
	"dagger/goreleaser/internal/dagger"
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
	Src *dagger.Directory,

	// Custom container to use as a base container. Must have 'goreleaser' available on PATH.
	// +optional
	Container *dagger.Container,

	// Version (image tag) to use as a goreleaser binary source.
	// +optional
	// +default="latest"
	Version string,

	// Configuration file.
	// +optional
	Config *dagger.File,

	// Mount netrc credentials for a private git repository.
	// +optional
	Netrc *dagger.Secret,

	// Disable mounting cache volumes.
	//
	// +optional
	DisableCache bool,
) *Goreleaser {
	if Container == nil {
		Container = defaultContainer(Version)
	}

	flags := []string{"goreleaser"}
	srcDir := "/work/src"
	Container = Container.With(
		func(c *dagger.Container) *dagger.Container {
			if Config != nil {
				cfgPath, err := Config.Name(ctx)
				if err != nil {
					panic(fmt.Errorf("resolving configuration file name: %w", err))
				}
				c = c.WithMountedFile(cfgPath, Config)
				flags = append(flags, "--config", cfgPath)
			}
			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if !DisableCache {
				c = withGoModuleCacheFn(dag.CacheVolume("go-mod"), nil, "")(c)
				c = withGoBuildCacheFn(dag.CacheVolume("go-build"), nil, "")(c)
			}
			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if Netrc != nil {
				c = c.WithMountedSecret("/root/.netrc", Netrc)
			}
			return c
		}).
		WithWorkdir(srcDir).
		WithMountedDirectory(srcDir, Src)

	gr := &Goreleaser{
		Container:      Container,
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
	expect := dagger.ReturnTypeSuccess
	if ignoreError {
		expect = dagger.ReturnTypeAny
	}
	return gr.Container.WithExec(append([]string{"goreleaser"}, args...), dagger.ContainerWithExecOpts{Expect: expect}).Stdout(ctx)
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
