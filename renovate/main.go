// A Module to run renovate against a remote project to check for any dependency updates.
// This will attempt create Pull/Merge Requests fodepending on platform provided, in ex. github.

package main

import (
	"context"
	"dagger/renovate/internal/dagger"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

// Renovate tasks
type Renovate struct {
	// +private
	Project string

	// +private
	EndpointURL string

	// +private
	Platform string

	// +private
	Base *dagger.Container

	// +private
	Token *dagger.Secret

	// +private
	Auths []Auth

	// +private
	Secrets []Secret

	// +private
	GitPrivateKey *dagger.Secret

	// +private
	Author string

	// +private
	Email string

	// +private
	EnabledManagers string
}

type Auth struct {
	Hostname string
	Username string
	Password *dagger.Secret
}

type Secret struct {
	Name  string
	Value *dagger.Secret
}

const globalExtends = `
[
	"config:recommended",
	":semanticCommitTypeAll(fix)",
	":prHourlyLimitNone",
	":prConcurrentLimit20",
	":disableDependencyDashboard",
	"regexManagers:dockerfileVersions",
	"regexManagers:gitlabPipelineVersions",
	"regexManagers:helmChartYamlAppVersions"
]
`

//go:embed renovate-managers.json
var customManagers string

func New(
	// repo project slug
	project string,

	// Gitlab API token to the repo being renovated
	token *dagger.Secret,

	// Endpoint URL for example https://hostname/api/v4
	endpointURL string,

	// set platform for renovate to use. in ex. "gitlab"
	// +optional
	// +default="gitlab"
	platform string,

	// renovate base image
	// +optional
	base *dagger.Container,

	// private git key for signing commits
	// note: Renovate does not support password protected keys
	// +optional
	gitPrivateKey *dagger.Secret,

	// git author for creating branches/commits
	// +optional
	// +default="RenovateBot"
	author string,

	// git email for creating branches/commits
	// +optional
	// +default="bot@example.com"
	email string,

	// +optional
	// +default=""
	enabledManagers string,
) *Renovate {
	if base == nil {
		base = dag.Container().From("renovate/renovate:41.23.5-full")
	}
	return &Renovate{
		Project:         project,
		Base:            base,
		Token:           token,
		EndpointURL:     endpointURL,
		Platform:        platform,
		GitPrivateKey:   gitPrivateKey,
		Author:          author,
		Email:           email,
		EnabledManagers: enabledManagers,
	}
}

// Add authentication to a OCI registry
func (m *Renovate) WithRegistryAuth(
	// registry's hostname
	hostname string,
	// username in registry
	username string,
	// password or token for registry
	password *dagger.Secret,
) *Renovate {
	m.Auths = append(m.Auths, Auth{
		Hostname: hostname,
		Username: username,
		Password: password,
	})
	return m
}

// Add a renovate secret.
// Can we referenced as "{{ secrets.MY_SECRET_NAME }}" in other renovate config.
func (m *Renovate) WithSecret(
	// name of the secret
	name string,
	// value of the secret
	value *dagger.Secret,
) *Renovate {
	m.Secrets = append(m.Secrets, Secret{
		Name:  name,
		Value: value,
	})
	return m
}

func (m *Renovate) getHostRules(ctx context.Context) (*dagger.Secret, error) {
	type hostRule struct {
		MatchHost string `json:"matchHost"`
		HostType  string `json:"hostType"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}

	hostRules := make([]hostRule, len(m.Auths))
	for i, auth := range m.Auths {
		registryPasswordText, err := auth.Password.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get registry password's plaintext: %w", err)
		}

		hostRules[i] = hostRule{
			MatchHost: auth.Hostname,
			HostType:  "docker",
			Username:  auth.Username,
			Password:  registryPasswordText,
		}
	}

	hostRulesJson, err := json.Marshal(hostRules)
	if err != nil {
		return nil, err
	}

	// TODO RegistryConfig uses the sha256 digest of the value as the name of the secret
	return dag.SetSecret("renovate-host-rules", string(hostRulesJson)), nil
}

func (m *Renovate) getSecrets(ctx context.Context) (*dagger.Secret, error) {

	secretsMap := make(map[string]string, len(m.Secrets))
	for _, s := range m.Secrets {
		plaintext, err := s.Value.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret value in plaintext: %w", err)
		}

		secretsMap[s.Name] = plaintext
	}

	secretsJson, err := json.Marshal(secretsMap)
	if err != nil {
		return nil, err
	}

	// TODO RegistryConfig uses the sha256 digest of the value as the name of the secret
	return dag.SetSecret("renovate-secrets", string(secretsJson)), nil
}

// Run renovate to update dependencies on the remote repository
func (m *Renovate) Update(ctx context.Context) (string, error) {
	// const author = "Renovate Bot"
	// const email = "bot@example.com"

	hostRules, err := m.getHostRules(ctx)
	if err != nil {
		return "", err
	}

	secrets, err := m.getSecrets(ctx)
	if err != nil {
		return "", err
	}

	return m.Base.
		WithEnvVariable("RENOVATE_ENDPOINT", m.EndpointURL).
		WithEnvVariable("RENOVATE_PLATFORM", m.Platform).
		WithSecretVariable("RENOVATE_TOKEN", m.Token).
		WithEnvVariable("RENOVATE_USERNAME", "renovate-bot").
		WithEnvVariable("RENOVATE_AUTODISCOVER", "false").
		WithEnvVariable("RENOVATE_GLOBAL_EXTENDS", globalExtends).
		WithEnvVariable("RENOVATE_ALLOWED_POST_UPGRADE_COMMANDS", `["^.*$"]`).
		WithSecretVariable("RENOVATE_HOST_RULES", hostRules).
		// WithEnvVariable("GIT_AUTHOR_NAME", author).
		// WithEnvVariable("GIT_AUTHOR_EMAIL", email).
		// WithEnvVariable("GIT_COMMITTER_NAME", author).
		// WithEnvVariable("GIT_COMMITTER_EMAIL", email).
		WithEnvVariable("RENOVATE_GIT_AUTHOR", fmt.Sprintf("%s <%s>", m.Author, m.Email)).
		With(func(c *dagger.Container) *dagger.Container {
			if m.GitPrivateKey != nil {
				return c.WithSecretVariable("RENOVATE_GIT_PRIVATE_KEY", m.GitPrivateKey)
			}

			return c
		}).
		WithEnvVariable("GPG_TTY", "$(tty)").
		// WithEnvVariable("RENOVATE_GIT_IGNORED_AUTHORS", email).
		WithEnvVariable("RENOVATE_REQUIRE_CONFIG", "optional").
		WithEnvVariable("RENOVATE_ONBOARDING", "false").
		WithEnvVariable("RENOVATE_ENABLED_MANAGERS", m.EnabledManagers).
		WithEnvVariable("RENOVATE_CUSTOM_MANAGERS", customManagers).
		WithSecretVariable("RENOVATE_SECRETS", secrets).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		// WithMountedSecret("/home/ubuntu/.docker/config.json", m.RegistryConfig.Secret()).
		// WithEnvVariable("HELM_REGISTRY_CONFIG", "/root/.docker/config.json").
		WithEnvVariable("LOG_LEVEL", "debug").
		// Terminal(dagger.ContainerTerminalOpts{Cmd: []string{"bash"}}).
		// We could use --platform=local to use the local source repo.
		WithExec([]string{"renovate", m.Project}).
		Stdout(ctx)

	/*
	  The error from OpenTelemetry is because OTEL_EXPORTER_OTLP_ENDPOINT env is set by Dagger and renovate used OpenTelemetry https://docs.renovatebot.com/opentelemetry/ so it tries to publish telemetroy to Dagger's OTEL stuff and fails (for an unknown reason).  The error is not fatal.
	*/
}
