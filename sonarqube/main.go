// Sonarqube module for local development use/scanning ONLY.

package main

import (
	"context"
	"dagger/sonarqube/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type Sonarqube struct{}

const hardcodedPassword = "MyHardcodedPass123!"

// start up sonar-server as a service
func (m *Sonarqube) Service() *dagger.Service {

	return dag.Container().
		From("sonarqube:community").
		WithEnvVariable("SONAR_ES_BOOTSTRAP_CHECKS_DISABLE", "true").
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
		).AsService()
}

// scan a source directory with sonar-scanner and get a report from sonar-server
func (m *Sonarqube) Scan(ctx context.Context,
	projectName string,
	src *dagger.Directory) (*dagger.File, error) {
	//start sonar-server
	sonarSvc, err := m.Service().Start(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to start sonar service: %w", err)
	}

	//setup admin pw/project
	if err := m.serverSetup(ctx, sonarSvc, projectName); err != nil {
		return nil, err
	}

	//generate sonar token
	sonarToken, err := m.generateSonarToken(ctx, sonarSvc, projectName)

	if err != nil {
		return nil, err
	}

	// run sonar scan
	_, err = dag.Container().
		From("sonarsource/sonar-scanner-cli:latest").
		WithServiceBinding("sonar-server", sonarSvc).
		WithDirectory("/src", src, dagger.ContainerWithDirectoryOpts{Owner: "scanner-cli"}).
		WithWorkdir("/src").
		WithSecretVariable("SONAR_TOKEN", sonarToken).
		WithEnvVariable("SONAR_HOST_URL", "http://sonar-server:9000").
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{
			"sonar-scanner",
			"-Dsonar.projectName=" + projectName,
			"-Dsonar.projectKey=" + projectName,
			"-Dsonar.sources=.",
			"-Dsonar.tests=.",
			"-Dsonar.test.inclusions=src/**/*.test.*,src/**/*.spec.*",
			"-Dsonar.exclusions=node_modules/**,dist/**,build/**,coverage/**,public/**,src/api/**",
		}).Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to run sonarqube scan: %w", err)
	}

	// get json report of issues
	report := m.getReport(sonarSvc, sonarToken, projectName)

	return report, nil
}

func (m *Sonarqube) curlCtr(svc *dagger.Service) *dagger.Container {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithServiceBinding("sonar-server", svc).
		WithEnvVariable("CACHEBUSTER", time.Now().String())
}

// create sonar project token to run a scan with
func (m *Sonarqube) generateSonarToken(ctx context.Context, svc *dagger.Service, projectName string) (*dagger.Secret, error) {

	token, err := m.curlCtr(svc).
		WithExec([]string{"sh", "-c", fmt.Sprintf(`
		curl -s -X POST -u admin:%s http://sonar-server:9000/api/user_tokens/generate \
			-d "name=dagger-token" \
			-d "type=PROJECT_ANALYSIS_TOKEN" \
			-d "projectKey=%s" | jq -r '.token'
	`, hardcodedPassword, projectName)}).Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to automatically create sonar token: %w", err)
	}

	return dag.SetSecret("SONAR_TOKEN", strings.TrimSpace(token)), nil

}

// get generated report in sonar
func (m *Sonarqube) getReport(svc *dagger.Service, token *dagger.Secret, projectName string) *dagger.File {

	return m.curlCtr(svc).
		WithEnvVariable("PROJECT_NAME", projectName).
		WithSecretVariable("SONAR_TOKEN", token).
		WithExec([]string{"sh", "-c", "sleep 20"}). //HACK: Sonar takes time to build the report after scan, so we wait for it to finish
		WithExec([]string{
			"sh",
			"-c",
			`curl -u "$SONAR_TOKEN:" \
      "http://sonar-server:9000/api/issues/search?componentKeys=$PROJECT_NAME"`,
		}, dagger.ContainerWithExecOpts{RedirectStdout: "sonar-report.json"}).File("sonar-report.json")

}

// create initial admin password and create new project with project name
func (m *Sonarqube) serverSetup(ctx context.Context, svc *dagger.Service, projectName string) error {
	// change admin password on first run
	_, err := m.curlCtr(svc).
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
		return fmt.Errorf("failed to automatically create admin password: %w", err)
	}

	// create project in sonarqube with project name
	_, err = m.curlCtr(svc).
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
		return fmt.Errorf("failed to automatically create project: %w", err)
	}

	return nil

}
