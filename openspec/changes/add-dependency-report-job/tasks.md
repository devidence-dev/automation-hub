# Tasks

## 1. Pre-checks

- [x] 1.1 Confirm that `add-workflow-scheduler` is applied (scheduler, dispatch core and `schedule` validation are on the branch). Verify that `go test ./internal/services/scheduler/...` passes.
- [x] 1.2 Check the following against Renovate `44.151.0` source: the `updateType` values, that `level >= 50` means error, the report shape (`repositories`, `packageFiles`, `libYearsWithStatus`, `problems`), the `RENOVATE_REPOSITORIES` env format, and the image entrypoint and `sh`. Record the findings in the ChangePlan devlog. Verify that the synthetic fixture in 4.1 matches them.
- [x] 1.3 Check the `cronjob.kubernetes.io/instantiate: manual` annotation and the ownerReference behavior in kubectl's `pkg/cmd/create/create_job.go`. Verify by quoting the source lines in the PR or devlog.

## 2. Config and wiring

- [x] 2.1 Add `JobCommandConfig` (`jobs:`) and `DependencyReportConfig` (`dependency_report:` with `webhook_token`, `telegram_chat_id`, `forks []string`) to `internal/config`, and extend command validation to `workflows`, `restarts` and `jobs` together, plus `schedule` on jobs. Verify with table tests in `config_test.go`: a duplicate across types is rejected, an invalid job schedule is rejected, a valid jobs-only config is accepted, and `forks` given as a string (a quoted reference) is rejected.
- [x] 2.2 In `main.go`, include `len(cfg.Jobs) > 0` in the bot condition and the `kube.NewInClusterClient` condition, register job commands with `SetMyCommands`, and register `POST /webhook/dependency-report` only when `dependency_report:` is set. Verify by starting the binary in Docker with a jobs-only config and checking the logs. Also check that the route returns 404 when the section is absent (handler test using the same router setup helper, if one is extracted). _Done with `TestRegisterRoutesDependencyReport` (extracted `registerRoutes`) plus `go build ./...`; the binary was not started in Docker because it needs the in-cluster ServiceAccount._
- [x] 2.3 Update `configs/config.yaml.example` and the README with `dependency_report:` (Infisical references; `forks: ${DEPENDENCY_REPORT_FORKS}` unquoted, with a comment explaining why) and a `jobs:` entry (`run_dependency_report`, `namespace: internal-apps`, `cronjob: dependency-report`, `schedule: "0 9 * * 1"`). Verify that the example loads in the config tests.

## 3. Job commands

- [x] 3.1 Implement `CreateJobFromCronJob` in `internal/services/kube`. Verify with an `httptest` test: the GET and POST paths, the Job name `<cronjob>-manual-<unix>`, the labels and annotations copied, the manual annotation present, and an error returned on 403.
- [x] 3.2 Add the job command type to `BotHandler` and to the scheduler registration, with the replies `🚀 Job *<name>* created from *<cronjob>*.` and `❌ Failed to create the job. Check the hub logs or try again later.`. Verify with bot handler tests for success, API error and unauthorized chat, and a scheduler registration test with a fake creator.

## 4. Dependency report

- [x] 4.1 Add a synthetic Renovate report fixture under `internal/handlers/testdata/` covering major, minor, digest-only, a fork, more than 40 deps, `libYearsWithStatus`, and problems with level 40 and 50. Verify that it parses with `parseReport`.
- [x] 4.2 Implement `buildMessages(report, forks)`: per-repo lines, 🔴/🟡, dedup, header and fork tag, `outdated/total · libyears`, `⚠️ N errores en el scan`, the 40-line cap with `… y N más`, "todo al día", and HTML escaping with line-safe 4096 truncation. Verify with table tests, one per scenario in `specs/dependency-report/spec.md`.
- [x] 4.3 Implement the Renovate failure message (`❌ Renovate falló (exit N) en el pod <pod>` with an escaped `<pre>` tail that fits in 4096). Verify with a test that uses a tail containing `<html>` and a 10 KB tail.
- [x] 4.4 Add `SendHTML` to `internal/services/telegram`, sharing the retry loop with `SendMessage`. Verify with a client test against the existing fake endpoint that it sends `parse_mode=HTML` and that `SendMessage` still sends `Markdown`.
- [x] 4.5 Implement the HTTP handler in `internal/handlers/dependency_report.go`: constant-time token check (401), `MaxBytesReader`, multipart validation (400), 202, and a background send with its own timeout to `dependency_report.telegram_chat_id`. Verify with `httptest` tests: no token gives 401 and no send; a non-integer `exit_code` gives 400; forks come from config, not from the request; a valid request gives 202 and the fake messenger receives the expected messages; Telegram failure is logged and nothing panics.

## 5. Cross-repo

- [x] 5.1 (user) Create the PAT and, in automation-hub's Infisical folder (next to `CONFIG_YAML`), add `RENOVATE_TOKEN`, `DEPENDENCY_REPORT_REPOSITORIES`, `DEPENDENCY_REPORT_FORKS` and `DEPENDENCY_REPORT_TOKEN`. Verify that both ExternalSecrets sync.
- [x] 5.2 (homelab) In `kubernetes/internal-apps/automation-hub/`, add `dependency-report.yaml` with the ExternalSecret (store `infisical-automation-hub`) and the CronJob `dependency-report` (`RENOVATE_REPOSITORIES` from `DEPENDENCY_REPORT_REPOSITORIES`, `suspend: true`, label `app.kubernetes.io/name: dependency-report`, Renovate initContainer that always exits 0, a pinned reporter image, `automountServiceAccountToken: false`, emptyDir `/out`, limits from the ChangePlan), and add it to `directory.include` in `application.yaml`. Verify with `kubectl apply --dry-run=server` and that Argo shows it under the automation-hub app.
- [x] 5.3 (homelab) Add both NetworkPolicies (egress from `dependency-report` to automation-hub:8080, and ingress on automation-hub from `dependency-report`:8080). Verify on the first run that the reporter's curl gets 202.
- [x] 5.4 (homelab) Add Role `dependency-report-trigger` (cronjobs `get` with `resourceNames: [dependency-report]`, jobs `create`) and a RoleBinding to the automation-hub ServiceAccount. Verify with `kubectl auth can-i create jobs -n internal-apps --as=system:serviceaccount:internal-apps:automation-hub`.
- [ ] 5.5 (user) Add `dependency_report:` and the `jobs:` entry to `CONFIG_YAML`, deploy, and run `/run_dependency_report`. Verify that the Telegram messages match the format, and that the API shows no new issues, PRs or branches in the scanned repos.
- [ ] 5.6 (user) Test the failure path with a non-existent repo in `DEPENDENCY_REPORT_REPOSITORIES`, then revert. Verify that the `❌ Renovate falló` message arrives.
- [x] 5.7 (workflows) Delete `ops-dependency-report.yml` and its README row and section. Verify with `actionlint` on the remaining workflows.

## 6. Validation

- [x] 6.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. The task passes only when both commands succeed on the current code.
