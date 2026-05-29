// A generated module for Sonarqube functions

package main

import (
	"context"
	"dagger/sonarqube/internal/dagger"
	"fmt"
	"time"
)

type Sonarqube struct{}

func (m *Sonarqube) Scan(ctx context.Context,
	projectName string) (string, error) {
	hardcodedPassword := "MyHardcodedPass123!"

	sonarCtr := dag.Container().
		From("sonarqube:community").
		WithExposedPort(9000).
		WithDockerHealthcheck(
			[]string{
				"sh",
				"-c",
				`curl -sf http://localhost:9000/api/system/status | grep -q '"status":"UP"'`,
			},
			dagger.ContainerWithDockerHealthcheckOpts{
				Interval: "5s",
				Timeout:  "3s",
				Retries:  15,
			},
		)

	sonarSvc := sonarCtr.AsService()

	curlCtr := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithServiceBinding("sonar-server", sonarSvc).
		WithEnvVariable("CACHEBUSTER", time.Now().String())

	// set admin PW on first run
	_, err := curlCtr.
		WithExec([]string{
			"curl",
			"-X", "POST",
			"-u", "admin:admin", // Default credentials
			"-d", "login=admin",
			"-d", "previousPassword=admin",
			"-d", fmt.Sprintf("password=%s", hardcodedPassword),
			"http://sonar-server:9000/api/users/change_password",
		}).
		Sync(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to automatically change admin password: %w", err)
	}

	// create project
	_, err = curlCtr.
		WithExec([]string{
			"curl",
			"-X", "POST",
			"-u", fmt.Sprintf("admin:%s", hardcodedPassword),
			"-d", fmt.Sprintf("project=%s", projectName),
			"-d", fmt.Sprintf("name=%s", projectName),
			"http://sonar-server:9000/api/projects/create",
		}).
		Sync(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to automatically create project: %w", err)
	}
	// create project token
	projectToken, err := curlCtr.WithExec([]string{"sh", "-c", fmt.Sprintf(`
		curl -s -X POST -u admin:%s http://sonar-server:9000/api/user_tokens/generate \
			-d "name=dagger-token" \
			-d "type=PROJECT_ANALYSIS_TOKEN" \
			-d "projectKey=%s" | jq -r '.token'
	`, hardcodedPassword, projectName)}).Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to obtain project token: %w", err)
	}

	dag.SetSecret("SONAR_TOKEN", projectToken)

	return projectToken, nil
}
