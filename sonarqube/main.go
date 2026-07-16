// Sonarqube module for local development use/scanning ONLY.
// This module will start a sonar server as a service,
// run a scan with sonar-scanner CLI with a given source directory,
// and return a json report of any issues found.

package main

import (
	"context"
	"crypto/rand"
	"dagger/sonarqube/internal/dagger"
	"encoding/json"
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

type SonarIssue struct {
	Key       string `json:"key"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Component string `json:"component"`
	Line      int    `json:"line"`
}

type SonarPage struct {
	Issues []SonarIssue `json:"issues"`
}

type SonarServer struct {
	// The live running service
	Service *dagger.Service
	// The generated user token used for running scans
	SonarToken *dagger.Secret
	// The plaintext admin password (for logging into the UI)
	AdminToken *dagger.Secret
}

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

// starts sonar as a service and bootstraps it with an admin password, project, and sonar token for scanning
func (m *Sonarqube) Bootstrap(ctx context.Context) (*SonarServer, error) {

	sonarSvc, err := m.Service().Start(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to start sonar service: %w", err)
	}

	// generate admin PW for first use
	adminToken := m.generateRandomTokenAsSecret()
	adminPlaintext, err := adminToken.Plaintext(ctx)

	// setup admin pw and project in sonar
	if err := m.serverSetup(ctx, sonarSvc, adminToken); err != nil {
		return nil, err
	}

	// generate sonar token
	sonarToken, err := m.generateSonarToken(ctx, sonarSvc, adminToken)

	// 5. Print out the login details for UI
	fmt.Println("\n====================================================")
	fmt.Println("🚀  SONARQUBE LOCAL ENGINE STARTED SUCCESSFULLY!  🚀")
	fmt.Println("====================================================")
	fmt.Println("  URL:      http://localhost:9000")
	fmt.Println("  Username: admin")
	fmt.Printf("  Password: %s\n", adminPlaintext)
	fmt.Println("====================================================")

	return &SonarServer{
		Service:    sonarSvc,
		SonarToken: sonarToken,
		AdminToken: adminToken,
	}, nil
}

// scan a source directory with sonar-scanner and get a report from sonar-server
func (m *Sonarqube) Scan(ctx context.Context,
	// +defaultPath="/"
	src *dagger.Directory,
	// comma separated list of impact severities to use when generating report
	// +optional
	// +default="MEDIUM,HIGH"
	impactSeverities string) (string, error) {
	// // start sonar-server
	// sonarSvc, err := m.Service().Start(ctx)

	// if err != nil {
	// 	return nil, fmt.Errorf("failed to start sonar service: %w", err)
	// }

	// // defer sonarSvc.Stop(ctx)

	// // change admin pw on first use
	// adminToken := m.generateRandomTokenAsSecret()

	// // setup admin pw and project in sonar
	// if err := m.serverSetup(ctx, sonarSvc, adminToken); err != nil {
	// 	return nil, err
	// }

	// // generate sonar token
	// sonarToken, err := m.generateSonarToken(ctx, sonarSvc, adminToken)

	server, err := m.Bootstrap(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initialize sonarqube: %w", err)
	}

	// run sonar scan
	_, err = dag.Container().
		From("sonarsource/sonar-scanner-cli:latest").
		WithServiceBinding("sonar-server", server.Service).
		WithDirectory("/src", src, dagger.ContainerWithDirectoryOpts{Owner: "scanner-cli"}).
		WithWorkdir("/src").
		WithSecretVariable("SONAR_TOKEN", server.SonarToken).
		WithEnvVariable("SONAR_HOST_URL", "http://sonar-server:9000").
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{
			"sonar-scanner",
			"-Dsonar.projectName=proj1",
			"-Dsonar.projectKey=proj1",
			"-Dsonar.qualitygate.wait=true",
		}).Sync(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to run sonarqube scan: %w", err)
	}

	// get json report of issues
	report := m.getReport(server.Service, server.SonarToken, impactSeverities)

	return m.AnalyzeSonarReport(ctx, report)
}

// generate random sonar admin pw that passes sonarqube pw requirements
func (m *Sonarqube) generateRandomTokenAsSecret() *dagger.Secret {
	b := make([]byte, 6)
	_, _ = rand.Read(b)

	// %x forces lowercase and numbers. "A!" guarantees uppercase and special char.
	compliantPassword := fmt.Sprintf("A!%x", b)

	return dag.SetSecret("SONAR_ADMIN_TOKEN", compliantPassword)
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
func (m *Sonarqube) getReport(svc *dagger.Service, token *dagger.Secret, impactSeverities string) *dagger.File {

	fetchScript := `
	PAGE_SIZE=499
	BASE_URL="http://sonar-server:9000/api/issues/search?components=proj1&additionalFields=comments&issueStatuses=OPEN,CONFIRMED&impactSeverities=$IMPACT_SEVERITIES&ps=$PAGE_SIZE"
	
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
		WithEnvVariable("IMPACT_SEVERITIES", impactSeverities).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
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

// Parses sonarqube JSON report into human readable format. Will return an error if issues are found HIGH or above.
func (m *Sonarqube) AnalyzeSonarReport(ctx context.Context, reportFile *dagger.File) (string, error) {
	reportRaw, err := reportFile.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read sonar report file contents: %w", err)
	}

	var totalIssues int
	var blockers, criticals, highs, mediums, lows []SonarIssue

	seenIssues := make(map[string]bool)

	categorizeIssue := func(issue SonarIssue) {
		if seenIssues[issue.Key] {
			return
		}
		seenIssues[issue.Key] = true
		totalIssues++

		switch strings.ToUpper(issue.Severity) {
		case "BLOCKER":
			blockers = append(blockers, issue)
		case "CRITICAL":
			criticals = append(criticals, issue)
		case "HIGH", "MAJOR":
			highs = append(highs, issue)
		case "MEDIUM", "MINOR":
			mediums = append(mediums, issue)
		default:
			lows = append(lows, issue)
		}
	}

	trimmedRaw := strings.TrimSpace(reportRaw)
	// Parse standard JSON and multi-line JSONL
	var singlePage SonarPage
	if err := json.Unmarshal([]byte(trimmedRaw), &singlePage); err == nil && len(singlePage.Issues) > 0 {
		for _, issue := range singlePage.Issues {
			categorizeIssue(issue)
		}
	} else {
		lines := strings.Split(trimmedRaw, "\n")
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" {
				continue
			}

			var page SonarPage
			if err := json.Unmarshal([]byte(trimmedLine), &page); err != nil {
				return "", fmt.Errorf("failed to parse JSON on line %d (verify format): %w", i+1, err)
			}

			for _, issue := range page.Issues {
				categorizeIssue(issue)
			}
		}
	}

	failCount := len(blockers) + len(criticals) + len(highs)

	// Use a strings.Builder to capture the entire report output
	var report strings.Builder

	fmt.Fprintf(&report, "\n=========================================\n")
	fmt.Fprintf(&report, "       SONARQUBE ANALYSIS SUMMARY        \n")
	fmt.Fprintf(&report, "=========================================\n")
	fmt.Fprintf(&report, "Total Unique Issues Found: %d\n", totalIssues)
	fmt.Fprintf(&report, "-----------------------------------------\n")
	fmt.Fprintf(&report, " 🛑 BLOCKER:  %d\n", len(blockers))
	fmt.Fprintf(&report, " 💥 CRITICAL: %d\n", len(criticals))
	fmt.Fprintf(&report, " ⚠️  HIGH:     %d\n", len(highs))
	fmt.Fprintf(&report, " 📝 MEDIUM:   %d\n", len(mediums))
	fmt.Fprintf(&report, " ℹ️  LOW:      %d\n", len(lows))
	fmt.Fprintf(&report, "=========================================\n")

	// Helper closure to build categorized blocks
	appendIssueGroup := func(title string, list []SonarIssue) {
		if len(list) == 0 {
			return
		}
		fmt.Fprintf(&report, "\n>>> %s ISSUES (%d) <<<\n", title, len(list))
		for _, issue := range list {
			fmt.Fprintf(&report, "  - [KEY: %s]\n", issue.Key)
			fmt.Fprintf(&report, "    Component: %s (Line: %d)\n", issue.Component, issue.Line)
			fmt.Fprintf(&report, "    Message:   %s\n\n", issue.Message)
		}
	}

	appendIssueGroup("BLOCKER", blockers)
	appendIssueGroup("CRITICAL", criticals)
	appendIssueGroup("HIGH", highs)
	appendIssueGroup("MEDIUM", mediums)

	fmt.Fprintf(&report, "=========================================\n")

	// return error and print report summary
	if failCount > 0 {
		// Output to standard out so it is captured in successful execution history if checked
		// fmt.Print(report.String())

		// Return the entire structured string inside the error.
		return "", fmt.Errorf("%s\nQuality Gate Failed: Found %d blocking issue(s) (Blocker: %d, Critical: %d, High: %d)",
			report.String(), failCount, len(blockers), len(criticals), len(highs))
	}

	successMsg := fmt.Sprintf("Success! Analyzed %d unique issues. Code quality thresholds passed.", totalIssues)
	fmt.Println(report.String())
	fmt.Println(successMsg)

	return successMsg, nil
}
