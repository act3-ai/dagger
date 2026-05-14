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
	Repositories []string

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
	LogLevel string
}

type Auth struct {
	Hostname string
	HostType string
	Username string
	Token    *dagger.Secret
}

type Secret struct {
	Name  string
	Value *dagger.Secret
}

const globalExtends = `
[
	"config:recommended",
	":dependencyDashboard",
	":semanticCommitTypeAll(fix)",
	":prHourlyLimitNone",
	":prConcurrentLimit20",
	"customManagers:dockerfileVersions",
	"customManagers:gitlabPipelineVersions",
	"customManagers:helmChartYamlAppVersions"
]
`

//go:embed renovate-managers.json
var customManagers string

// +cache="never"
func New(
	// project slug of repositories for renovate to scan.
	repositories []string,

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
	// renovate log level. refer to https://docs.renovatebot.com/troubleshooting/#log-debug-levels
	// +optional
	// +default="debug"
	logLevel string,
) *Renovate {
	if base == nil {
		base = dag.Container().From("renovate/renovate:full")
	}
	return &Renovate{
		Repositories:  repositories,
		Base:          base,
		Token:         token,
		EndpointURL:   endpointURL,
		Platform:      platform,
		GitPrivateKey: gitPrivateKey,
		Author:        author,
		Email:         email,
		LogLevel:      logLevel,
	}
}

// Add an env variable to the renovate container
func (m *Renovate) WithEnvVariable(
	// name of the secret variable
	name string,
	// value of the secret
	value string,
) *Renovate {
	m.Base = m.Base.WithEnvVariable(name, value)
	return m
}

// Add a secret env variable to the renovate container
func (m *Renovate) WithSecretVariable(
	// name of the secret variable
	name string,
	// value of the secret
	value *dagger.Secret,
) *Renovate {
	m.Base = m.Base.WithSecretVariable(name, value)
	return m
}

// Add a host rule to renovate for private package authentication.
// See: https://docs.renovatebot.com/configuration-options/#hostrules
func (m *Renovate) WithHostRule(
	// registry's hostname
	hostname string,
	// host type, in ex. "gitlab"
	// +optional
	// +default="docker"
	hostType string,
	// username in registry
	// +optional
	username string,
	// password or token for registry
	token *dagger.Secret,
) *Renovate {
	m.Auths = append(m.Auths, Auth{
		Hostname: hostname,
		HostType: hostType,
		Username: username,
		Token:    token,
	})
	return m
}

// Add a renovate secret.
// Can be referenced as "{{ secrets.MY_SECRET_NAME }}" in other renovate config.
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
		Token     string `json:"token,omitempty"`
		Username  string `json:"username,omitempty"`
		Password  string `json:"password,omitempty"`
	}

	hostRules := make([]hostRule, len(m.Auths))

	for i, auth := range m.Auths {
		tokenText, err := auth.Token.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth secret plaintext: %w", err)
		}

		rule := hostRule{
			MatchHost: auth.Hostname,
			HostType:  auth.HostType,
		}

		// username/password auth
		if auth.Username != "" {
			rule.Username = auth.Username
			rule.Password = tokenText
		} else {
			// token auth
			rule.Token = tokenText
		}

		hostRules[i] = rule
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

// +cache="never"
// Run renovate to update dependencies on the remote repository
func (m *Renovate) Update(ctx context.Context) (string, error) {

	cmd := append([]string{"renovate"}, m.Repositories...)

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
		WithEnvVariable("RENOVATE_CUSTOM_MANAGERS", customManagers).
		WithSecretVariable("RENOVATE_SECRETS", secrets).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithEnvVariable("LOG_LEVEL", m.LogLevel).
		// HACK: OTEL_EXPORTER_OTLP_ENDPOINT is set by dagger and causes renovate to error, so we unset it
		// We could use --platform=local to use the local source repo.
		WithExec(append([]string{
			"env",
			"OTEL_EXPORTER_OTLP_ENDPOINT=",
		}, cmd...)).
		Stdout(ctx)
}
