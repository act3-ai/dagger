// Sonarqube module for local development use/scanning ONLY.

package main

import (
	"context"
	"crypto/rand"
	"dagger/sonarqube/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type Sonarqube struct{}

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
				`curl -sf --noproxy "*" http://localhost:9000/api/system/status | grep -q '"status":"UP"'`,
			},
			dagger.ContainerWithDockerHealthcheckOpts{
				Interval: "5s",
				Timeout:  "3s",
				Retries:  30,
			},
		).AsService()
}

// scan a source directory with sonar-scanner and get a report from sonar-server
func (m *Sonarqube) Scan(ctx context.Context,
	src *dagger.Directory) (*dagger.File, error) {
	//start sonar-server
	sonarSvc, err := m.Service().Start(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to start sonar service: %w", err)
	}

	adminToken := m.generateRandomTokenAsSecret()

	//setup admin pw and project in sonar
	if err := m.serverSetup(ctx, sonarSvc, adminToken); err != nil {
		return nil, err
	}

	//generate sonar token
	sonarToken, err := m.generateSonarToken(ctx, sonarSvc, adminToken)

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
			"-Dsonar.projectName=proj1",
			"-Dsonar.projectKey=proj1",
		}).Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to run sonarqube scan: %w", err)
	}

	// get json report of issues
	report := m.getReport(sonarSvc, sonarToken)

	return report, nil
}

// generate random sonar admin pw and return as a secret
func (m *Sonarqube) generateRandomTokenAsSecret() *dagger.Secret {
	const (
		lower   = "abcdefghijklmnopqrstuvwxyz"
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers = "0123456789"
		special = "!@#%"
		all     = lower + upper + numbers + special
	)

	b := make([]byte, 16)

	// 1 guaranteed from each required set
	b[0] = lower[m.randomInt(len(lower))]
	b[1] = upper[m.randomInt(len(upper))]
	b[2] = numbers[m.randomInt(len(numbers))]
	b[3] = special[m.randomInt(len(special))]

	// fill remaining with full charset
	for i := 4; i < 16; i++ {
		b[i] = all[m.randomInt(len(all))]
	}

	// shuffle so pattern isn't predictable
	m.shuffle(b)

	token := string(b)

	return dag.SetSecret("SONAR_ADMIN_TOKEN", token)
}

// helper for token gen
func (m *Sonarqube) randomInt(max int) int {
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	return int(b[0]) % max
}

// helper for token gen
func (m *Sonarqube) shuffle(b []byte) {
	for i := len(b) - 1; i > 0; i-- {
		j := m.randomInt(i + 1)
		b[i], b[j] = b[j], b[i]
	}
}

// base curl container for sonar api queries
func (m *Sonarqube) curlCtr(svc *dagger.Service) *dagger.Container {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithServiceBinding("sonar-server", svc)
}

// create sonar project token to run a scan with
func (m *Sonarqube) generateSonarToken(ctx context.Context, svc *dagger.Service, adminToken *dagger.Secret) (*dagger.Secret, error) {

	token, err := m.curlCtr(svc).WithEnvVariable("CACHEBUSTER", time.Now().String()).WithSecretVariable("SONAR_ADMIN_TOKEN", adminToken).
		WithExec([]string{"sh", "-c",
			`http_code=$(curl -s --noproxy "*" -X POST -u admin:$SONAR_ADMIN_TOKEN \
                -d "name=dagger-token" \
                -d "type=PROJECT_ANALYSIS_TOKEN" \
                -d "projectKey=proj1" \
                -o /tmp/token_res \
                -w "%{http_code}" \
                http://sonar-server:9000/api/user_tokens/generate)

            if [ "$http_code" -ne 200 ]; then
                echo "HTTP Error $http_code: $(cat /tmp/token_res)"
                exit 1
            fi

            # extract token
            cat /tmp/token_res | jq -r '.token'
            `}).Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to automatically create sonar token: %w", err)
	}

	return dag.SetSecret("SONAR_TOKEN", strings.TrimSpace(token)), nil

}

// get generated report in sonar
func (m *Sonarqube) getReport(svc *dagger.Service, token *dagger.Secret) *dagger.File {

	return m.curlCtr(svc).
		WithSecretVariable("SONAR_TOKEN", token).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c", "sleep 20"}). //HACK: Sonar takes time to build the report after scan, so we wait for it to finish
		WithExec([]string{
			"sh",
			"-c",
			`curl --retry 5 --noproxy "*" -u "$SONAR_TOKEN:" \
      "http://sonar-server:9000/api/issues/search?componentKeys=proj1&impactSeverities=MEDIUM,HIGH"`,
		}, dagger.ContainerWithExecOpts{RedirectStdout: "sonar-report.json"}).File("sonar-report.json")

}

// create initial admin password and create new project with project name
func (m *Sonarqube) serverSetup(ctx context.Context, svc *dagger.Service, adminToken *dagger.Secret) error {

	curlCtr := m.curlCtr(svc).WithSecretVariable("SONAR_ADMIN_TOKEN", adminToken)
	// change admin password on first run
	adminPwOut, err := curlCtr.
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c",
			`http_code=$(curl -s --noproxy "*" -X POST -u admin:admin \
                -d "login=admin" \
                -d "previousPassword=admin" \
                -d "password=$SONAR_ADMIN_TOKEN" \
                -o /tmp/pw_res \
                -w "%{http_code}" \
                http://sonar-server:9000/api/users/change_password)

            if [ "$http_code" -ne 204 ]; then
                echo "HTTP Error $http_code: $(cat /tmp/pw_res)"
                exit 1
            fi
            `}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("sonarqube password change failed: %w\nDetails: %s", err, adminPwOut)
	}

	// create project in sonarqube with project name
	projectOut, err := curlCtr.
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c",
			`http_code=$(curl -s --noproxy "*" -X POST -u admin:$SONAR_ADMIN_TOKEN \
                -d "project=proj1" \
                -d "name=proj1" \
                -o /tmp/proj_res \
                -w "%{http_code}" \
                http://sonar-server:9000/api/projects/create)

            if [ "$http_code" -ne 200 ]; then
                echo "HTTP Error $http_code: $(cat /tmp/proj_res)"
                exit 1
            fi
            cat /tmp/proj_res
            `}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("sonarqube project creation failed: %w\nDetails: %s", err, projectOut)
	}

	return nil

}
