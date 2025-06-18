// A generated module for Docker functions
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
	"dagger/docker/internal/dagger"
	"encoding/json"
	"fmt"
)

type Docker struct {
	// +private
	Source *dagger.Directory
	// +private
	Secrets []Secret
	// +private
	RegistryCreds []RegistryCreds
	// +private
	BuildArg []dagger.BuildArg
	// +private
	Labels []Labels
	// +private
	PublishRef []string
}

type Secret struct {
	Name  string
	Value *dagger.Secret
}

type RegistryCreds struct {
	Registry string
	Username string
	Password *dagger.Secret
}

type BuildArgs struct {
	Name  string
	Value string
}

type Labels struct {
	Name  string
	Value string
}

func New(
	// top level source code directory
	// +optional
	src *dagger.Directory,
) *Docker {
	return &Docker{
		Source: src,
	}
}

// Add a docker secret to builds
func (d *Docker) WithSecret(
	// name of the secret
	name string,
	// value of the secret
	value *dagger.Secret,
) *Docker {
	d.Secrets = append(d.Secrets, Secret{
		Name:  name,
		Value: value,
	})
	return d
}

// Add docker registry creds to builds
func (d *Docker) WithRegistryCreds(
	// name of the registry
	registry string,
	// username for registry
	username string,
	// password for registry
	password *dagger.Secret,
) *Docker {
	d.RegistryCreds = append(d.RegistryCreds, RegistryCreds{
		Registry: registry,
		Username: username,
		Password: password,
	})
	return d
}

// Add docker registry creds to builds
func (d *Docker) WithDockerConfig(
	ctx context.Context,
	// file path to docker config json
	file *dagger.File,
) (*Docker, error) {

	// Read the contents of the dockerConfig
	configData, err := file.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker config: %w", err)
	}

	// Struct to parse json
	var config struct {
		Auths map[string]struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auths"`
	}

	// Parse the JSON
	if err := json.Unmarshal([]byte(configData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse docker config JSON: %w", err)
	}

	// Extract and append credentials
	for registry, creds := range config.Auths {
		daggerSecret := dag.SetSecret(registry, creds.Password)

		d.RegistryCreds = append(d.RegistryCreds, RegistryCreds{
			Registry: registry,
			Username: creds.Username,
			Password: daggerSecret,
		})
	}
	return d, err
}

// Add docker build args to builds
func (d *Docker) WithBuildArg(
	// name of the secret
	name string,
	// value of the secret
	value string,
) *Docker {
	d.BuildArg = append(d.BuildArg, dagger.BuildArg{
		Name:  name,
		Value: value,
	})
	return d
}

// Add labels to builds
func (d *Docker) WithLabel(
	// name of the secret
	name string,
	// value of the secret
	value string,
) *Docker {
	d.Labels = append(d.Labels, Labels{
		Name:  name,
		Value: value,
	})
	return d
}

// Retrieve secrets and set them in Dagger with dynamic names
func (d *Docker) getSecrets(ctx context.Context) ([]*dagger.Secret, error) {
	secretSlice := make([]*dagger.Secret, 0, len(d.Secrets))
	for _, s := range d.Secrets {
		plaintext, err := s.Value.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret value in plaintext for %s: %w", s.Name, err)
		}
		secret := dag.SetSecret(s.Name, plaintext)
		secretSlice = append(secretSlice, secret)
	}

	return secretSlice, nil
}

// Build image from Dockerfile
func (d *Docker) Build(
	ctx context.Context,
	// target stage of image build
	// +optional
	// +default="ci"
	target string,
	// platform to build with. value of [os]/[arch], example: linux/amd64, linux/arm64
	// +default="linux/amd64"
	platform dagger.Platform,
) (*dagger.Container, error) {

	//get secrets
	secrets, err := d.getSecrets(ctx)
	if err != nil {
		return nil, err
	}

	ctr := d.Source.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Target:    target,
		Secrets:   secrets,
		BuildArgs: d.BuildArg,
		Platform:  platform,
	})

	//Apply labels to container
	for _, label := range d.Labels {
		ctr = ctr.WithLabel(label.Name, label.Value)
	}

	//Apply registry authentication for each set of credentials
	for _, creds := range d.RegistryCreds {
		ctr = ctr.WithRegistryAuth(creds.Registry, creds.Username, creds.Password)
	}

	ctr, err = ctr.Sync(ctx)
	return ctr, err
}

// Build a multi-arch image index from Dockerfile and Publish to an OCI registry, returning a slice of image digest references.
func (d *Docker) Publish(ctx context.Context,
	// registry address to publish to, without tag
	address string,
	// comma separated list of tags to publish
	tags []string,
	// target stage of image build
	// +optional
	// +default="ci"
	target string,
	// platforms to build with. value of [os]/[arch], example: linux/amd64, linux/arm64
	// +default=["linux/amd64"]
	platforms []dagger.Platform,
) ([]string, error) {
	if len(tags) < 1 {
		return nil, fmt.Errorf("no tags provided, please specify a registry address and a set of tags")
	}

	//check for platforms and build each one
	platformVariants := make([]*dagger.Container, 0, len(platforms))
	for _, platform := range platforms {
		ctr, err := d.Build(ctx, target, platform)
		if err != nil {
			return nil, fmt.Errorf("building platform %s: %w", platform, err)
		}

		platformVariants = append(platformVariants, ctr)
	}

	// Publish tags to registry
	dgstAddrs := make([]string, 0, len(tags))
	for _, tag := range tags {
		addr := fmt.Sprintf("%s:%s", address, tag)
		a, err := dag.Container().Publish(ctx, addr,
			dagger.ContainerPublishOpts{
				PlatformVariants: platformVariants,
			})
		if err != nil {
			return nil, fmt.Errorf("publishing image index to %s: %w", addr, err)
		}
		dgstAddrs = append(dgstAddrs, a)
	}
	return dgstAddrs, nil
}
