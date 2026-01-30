package main

import (
	"context"
	"dagger/python/internal/dagger"
)

// publish python package/wheel
func (python *Python) Publish(ctx context.Context,
	// extra args to pass into uv build
	// +optional
	buildArgs []string,
	// url to publish python package/wheel to
	publishUrl string,
	// username for private publish url
	username string,
	// password for private publish url, must be given in dagger.Secret format, in ex. env://MY_TOKEN
	password *dagger.Secret,
) (*dagger.Container, error) {

	buildCmd := []string{"uv", "build"}
	buildCmd = append(buildCmd, buildArgs...)

	// Use the base image to avoid installing packages
	c := python.Base.
		WithEnvVariable("UV_PUBLISH_CHECK_URL", publishUrl+"/simple").
		WithEnvVariable("UV_PUBLISH_URL", publishUrl).
		WithEnvVariable("UV_PUBLISH_USERNAME", username).
		WithSecretVariable("UV_PUBLISH_PASSWORD", password).
		WithExec(buildCmd).
		WithExec(
			[]string{
				"uv",
				"publish",
			})

	return c, nil
}
