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
	publishUrl string,
	username string,
	password *dagger.Secret,
) (*dagger.Container, error) {

	buildCmd := []string{"uv", "build"}
	buildCmd = append(buildCmd, buildArgs...)

	c := python.Container().
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
