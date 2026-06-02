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

func (m *Sonarqube) curlCtr(svc *dagger.Service) *dagger.Container {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithServiceBinding("sonar-server", svc).
		WithEnvVariable("CACHEBUSTER", time.Now().String())
}

// start up sonar-server with the admin password changed and project created in sonar-server.
func (m *Sonarqube) Service(ctx context.Context,
	// name of project
	projectName string,
	// src directory to scan
	src *dagger.Directory) (*dagger.Service, error) {
	hardcodedPassword := "MyHardcodedPass123!"

	sonarSvc := dag.Container().
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

	curlCtr := m.curlCtr(sonarSvc)

	// change admin password on first run
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
		return nil, fmt.Errorf("failed to automatically change admin password: %w", err)
	}

	// create project in sonarqube
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
		return nil, fmt.Errorf("failed to automatically create project: %w", err)
	}

	return sonarSvc, nil
}

// scan a source directory with sonar-scanner and get a report from sonar-server
func (m *Sonarqube) Scan(ctx context.Context,
	projectName string,
	src *dagger.Directory) (*dagger.File, error) {
	sonarSvc, err := m.Service(ctx, projectName, src)
	if err != nil {
		return nil, fmt.Errorf("failed to automatically create project: %w", err)
	}

	// create sonar token needed to scan with
	projectToken, err := m.curlCtr(sonarSvc).
		WithExec([]string{"sh", "-c", fmt.Sprintf(`
		curl -s -X POST -u admin:%s http://sonar-server:9000/api/user_tokens/generate \
			-d "name=dagger-token" \
			-d "type=PROJECT_ANALYSIS_TOKEN" \
			-d "projectKey=%s" | jq -r '.token'
	`, hardcodedPassword, projectName)}).Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to obtain project token: %w", err)
	}

	projectToken = strings.TrimSpace(projectToken)
	sonarToken := dag.SetSecret("SONAR_TOKEN", projectToken)

	// run sonar scan
	_, err = dag.Container().
		From("sonarsource/sonar-scanner-cli:latest").
		WithServiceBinding("sonar-server", sonarSvc).
		WithDirectory("/src", src, dagger.ContainerWithDirectoryOpts{Owner: "scanner-cli"}).
		WithWorkdir("/src").
		WithSecretVariable("SONAR_TOKEN", sonarToken).
		WithEnvVariable("SONAR_HOST_URL", "http://sonar-server:9000").
		WithExec([]string{
			"sonar-scanner",
			"-Dsonar.projectName=" + projectName,
			"-Dsonar.projectKey=" + projectName,
			"-Dsonar.sources=.",
			"-Dsonar.tests=.",
			"-Dsonar.test.inclusions=src/**/*.test.*,src/**/*.spec.*",
			"-Dsonar.exclusions=node_modules/**,dist/**,build/**,coverage/**,public/**,src/api/**",
			// "-Dsonar.javascript.lcov.reportPaths=coverage/lcov.info",
			// "-Dsonar.issue.ignore.multicriteria=e1",
			// "-Dsonar.issue.ignore.multicriteria.e1.ruleKey=css:S4662",
			// "-Dsonar.issue.ignore.multicriteria.e1.resourceKey=**/*.css",
		}).Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to run sonarqube scan: %w", err)
	}

	// get json report of issues
	report := m.curlCtr(sonarSvc).
		WithSecretVariable("SONAR_TOKEN", sonarToken).
		// WithExec([]string{"sh", "-c", "sleep 20"}).
		WithExec([]string{
			"sh",
			"-c",
			`curl -u "$SONAR_TOKEN:" \
      "http://sonar-server:9000/api/issues/search?componentKeys=paul"`,
		}, dagger.ContainerWithExecOpts{RedirectStdout: "sonar-report.json"}).File("sonar-report.json")

	return report, nil
}
