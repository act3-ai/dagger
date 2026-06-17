// Sonarqube module for local development use/scanning ONLY.
// This module will start a sonar server as a service, run a scan with sonar-scanner against it,
// and return a json report of any issues found.

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

const changePasswordScript = `
http_code=$(curl -s --retry 5 --retry-delay 2 --noproxy "*" -X POST -u admin:admin \
	--data-urlencode "login=admin" \
	--data-urlencode "previousPassword=admin" \
	--data-urlencode "password=$SONAR_ADMIN_TOKEN" \
	-o /tmp/pw_res \
	-w "%{http_code}" \
	http://sonar-server:9000/api/users/change_password)

if [ "$http_code" -ne 204 ]; then
	echo "HTTP Error $http_code: $(cat /tmp/pw_res)"
	exit 1
fi
`

const createProjectScript = `
http_code=$(curl -s --retry 5 --retry-delay 2 --noproxy "*" -X POST -u admin:$SONAR_ADMIN_TOKEN \
	-d "project=proj1" \
	-d "name=proj1" \
	-o /tmp/proj_res \
	-w "%{http_code}" \
	http://sonar-server:9000/api/projects/create)

if [ "$http_code" -ne 200 ]; then
	echo "HTTP Error $http_code: $(cat /tmp/proj_res)"
	exit 1
fi
`

const generateTokenScript = `
http_code=$(curl -s --retry 5 --retry-delay 2 --noproxy "*" -X POST -u admin:$SONAR_ADMIN_TOKEN \
	-d "name=dagger-token" \
	-d "type=USER_TOKEN" \
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
`

const pollScript = `
	echo "Polling SonarQube until analysis processing completes..." >&2
	while true; do
		res=$(curl -s --noproxy "*" -u "$SONAR_TOKEN:" "http://sonar-server:9000/api/ce/component?component=proj1")
		queue_length=$(echo "$res" | jq '.queue | length')
		current_status=$(echo "$res" | jq -r '.current.status // "NONE"')
		
		echo "Current status: $current_status | Tasks in queue: $queue_length" >&2
		if [ "$queue_length" -eq 0 ] && [ "$current_status" = "SUCCESS" ]; then
			echo "Analysis complete and successful!" >&2
			break
		fi
		sleep 2
	done
	sleep 2 # Cooldown for index stabilization
	`

// start up sonar-server as a service
func (m *Sonarqube) Service() *dagger.Service {

	return dag.Container().
		From("sonarqube:community").
		WithEnvVariable("SONAR_ES_BOOTSTRAP_CHECKS_DISABLE", "true").
		// Define a system passcode that skips user authentication for health metrics
		WithEnvVariable("SONAR_WEB_SYSTEMPASSCODE", "dagger-health-token").
		WithExposedPort(9000).
		WithDockerHealthcheck(
			[]string{
				"sh",
				"-c",
				`curl -sf --noproxy "*" -H "X-Sonar-Passcode: dagger-health-token" http://localhost:9000/api/system/health | grep -q '"health":"GREEN"'`,
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
	// +defaultPath="/"
	src *dagger.Directory) (*dagger.File, error) {
	//start sonar-server
	sonarSvc, err := m.Service().Start(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to start sonar service: %w", err)
	}

	// defer sonarSvc.Stop(ctx)

	// change admin pw on first use
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

	token, err := m.curlCtr(svc).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithSecretVariable("SONAR_ADMIN_TOKEN", adminToken).
		WithExec([]string{"sh", "-c", generateTokenScript}).Stdout(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to automatically create sonar token: %w", err)
	}

	return dag.SetSecret("SONAR_TOKEN", strings.TrimSpace(token)), nil

}

// get generated report in sonar
func (m *Sonarqube) getReport(svc *dagger.Service, token *dagger.Secret) *dagger.File {

	fetchScript := `
	PAGE_SIZE=100
	BASE_URL="http://sonar-server:9000/api/issues/search?components=proj1&impactSeverities=LOW,MEDIUM,HIGH&ps=$PAGE_SIZE"
	
	echo "Fetching initial page..." >&2
	first_page=$(curl -s --retry 5 --noproxy "*" -u "$SONAR_TOKEN:" "${BASE_URL}&p=1")
	total_issues=$(echo "$first_page" | jq '.paging.total // 0')
	
	# Write the first page as a compressed single line (-c flag minimizes it to 1 line)
	echo "$first_page" | jq -c '.' > /tmp/report.jsonl

	total_pages=$(( (total_issues + PAGE_SIZE - 1) / PAGE_SIZE ))

	if [ "$total_pages" -gt 1 ]; then
		for p in $(seq 2 $total_pages); do
			echo "Fetching page $p of $total_pages..." >&2
			next_page=$(curl -s --retry 5 --noproxy "*" -u "$SONAR_TOKEN:" "${BASE_URL}&p=$p")
			
			# Append subsequent raw pages as single lines
			echo "$next_page" | jq -c '.' >> /tmp/report.jsonl
		done
	fi

	# Output the combined multi-document file
	cat /tmp/report.jsonl
	`

	return m.curlCtr(svc).
		WithSecretVariable("SONAR_TOKEN", token).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		// Wait for report generation to finish
		WithExec([]string{"sh", "-c", pollScript}).
		// Fetch report
		WithExec([]string{"sh", "-c", fetchScript}, dagger.ContainerWithExecOpts{
			RedirectStdout: "sonar-report.json",
		}).
		File("sonar-report.json")

}

// create initial admin password and create new project with project name
func (m *Sonarqube) serverSetup(ctx context.Context, svc *dagger.Service, adminToken *dagger.Secret) error {

	curlCtr := m.curlCtr(svc).WithSecretVariable("SONAR_ADMIN_TOKEN", adminToken)
	// change admin PW and create project in sonar-server
	Out, err := curlCtr.
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c", changePasswordScript}, dagger.ContainerWithExecOpts{}).
		WithExec([]string{"sh", "-c", createProjectScript}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("sonarqube server setup failed: %w\nDetails: %s", err, Out)
	}

	return nil
}
