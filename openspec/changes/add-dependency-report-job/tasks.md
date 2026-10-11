# Tasks

## 1. Pre-checks

- [ ] 1.1 Confirm that `add-workflow-scheduler` is applied (scheduler, dispatch core and `schedule` validation are on the branch). Verify that `go test ./internal/services/scheduler/...` passes.
- [ ] 1.2 Check the following against Renovate `44.151.0` source: the `updateType` values, that `level >= 50` means error, the report shape (`repositories`, `packageFiles`, `libYearsWithStatus`, `problems`), the `RENOVATE_REPOSITORIES` env format, and the image entrypoint and `sh`. Record the findings in the ChangePlan devlog. Verify that the synthetic fixture in 4.1 matches them.
- [ ] 1.3 Check the `cronjob.kubernetes.io/instantiate: manual` annotation and the ownerReference behavior in kubectl's `pkg/cmd/create/create_job.go`. Verify by quoting the source lines in the PR or devlog.

## 2. Config and wiring

- [ ] 2.1 Add `JobCommandConfig` (`jobs:`) and `DependencyReportConfig` (`dependency_report:`) to `internal/config`, and extend command validation to `workflows`, `restarts` and `jobs` together, plus `schedule` on jobs. Verify with table tests in `config_test.go`: a duplicate across types is rejected, an invalid job schedule is rejected, and a valid jobs-only config is accepted.
- [ ] 2.2 In `main.go`, include `len(cfg.Jobs) > 0` in the bot condition and the `kube.NewInClusterClient` condition, register job commands with `SetMyCommands`, and register `POST /webhook/dependency-report` only when `dependency_report:` is set. Verify by starting the binary in Docker with a jobs-only config and checking the logs. Also check that the route returns 404 when the section is absent (handler test using the same router setup helper, if one is extracted).
- [ ] 2.3 Update `configs/config.yaml.example` and the README with `dependency_report:` and a `jobs:` entry (`run_dependency_report`, `namespace: internal-apps`, `cronjob: dependency-report`, `schedule: "0 9 * * 1"`). Verify that the example loads in the config tests.

## 3. Job commands

- [ ] 3.1 Implement `CreateJobFromCronJob` in `internal/services/kube`. Verify with an `httptest` test: the GET and POST paths, the Job name `<cronjob>-manual-<unix>`, the labels and annotations copied, the manual annotation present, and an error returned on 403.
- [ ] 3.2 Add the job command type to `BotHandler` and to the scheduler registration, with the replies `Job *<name>* created from *<cronjob>*.` and `Failed to create the job. Check the hub logs or try again later.`. Verify with bot handler tests for success, API error and unauthorized chat, and a scheduler registration test with a fake creator.

## 4. Dependency report

- [ ] 4.1 Add a synthetic Renovate report fixture under `internal/handlers/testdata/` covering major, minor, digest-only, a fork, more than 40 deps, `libYearsWithStatus`, and problems with level 40 and 50. Verify that it parses with `parseReport`.
- [ ] 4.2 Implement `buildMessages(report, forks)`: per-repo lines, 🔴/🟡, dedup, header and fork tag, `outdated/total · libyears`, `⚠️ N errores en el scan`, the 40-line cap with `… y N más`, "todo al día", and HTML escaping with line-safe 4096 truncation. Verify with table tests, one per scenario in `specs/dependency-report/spec.md`.
- [ ] 4.3 Implement the Renovate failure message (`❌ Renovate falló (exit N) en el pod <pod>` with an escaped `<pre>` tail that fits in 4096). Verify with a test that uses a tail containing `<html>` and a 10 KB tail.
- [ ] 4.4 Add `SendHTML` to `internal/services/telegram`, sharing the retry loop with `SendMessage`. Verify with a client test against the existing fake endpoint that it sends `parse_mode=HTML` and that `SendMessage` still sends `Markdown`.
- [ ] 4.5 Implement the HTTP handler in `internal/handlers/dependency_report.go`: constant-time token check (401), `MaxBytesReader`, multipart validation (400), 202, and a background send with its own timeout to `dependency_report.telegram_chat_id`. Verify with `httptest` tests: no token gives 401 and no send; a bad `forks` gives 400; a valid request gives 202 and the fake messenger receives the expected messages; Telegram failure is logged and nothing panics.

## 5. Cross-repo

- [ ] 5.1 (user) Create the PAT and, in Infisical, the `/dependency-report` secrets (`RENOVATE_TOKEN`, `DEPENDENCY_REPORT_REPOSITORIES`, `DEPENDENCY_REPORT_FORKS`, `DEPENDENCY_REPORT_TOKEN`). Verify that the ExternalSecret syncs.
- [ ] 5.2 (homelab) Add the ExternalSecret and the CronJob `dependency-report` (`suspend: true`, label `app.kubernetes.io/name: dependency-report`, Renovate initContainer that always exits 0, a pinned reporter image, `automountServiceAccountToken: false`, emptyDir `/out`, limits from the ChangePlan). Verify with `kubectl apply --dry-run=server`.
- [ ] 5.3 (homelab) Add both NetworkPolicies (egress from `dependency-report` to automation-hub:8080, and ingress on automation-hub from `dependency-report`:8080). Verify on the first run that the reporter's curl gets 202.
- [ ] 5.4 (homelab) Add Role `dependency-report-trigger` (cronjobs `get` with `resourceNames: [dependency-report]`, jobs `create`) and a RoleBinding to the automation-hub ServiceAccount. Verify with `kubectl auth can-i create jobs -n internal-apps --as=system:serviceaccount:internal-apps:automation-hub`.
- [ ] 5.5 (user) Add `dependency_report:` and the `jobs:` entry to `CONFIG_YAML`, deploy, and run `/run_dependency_report`. Verify that the Telegram messages match the format, and that the API shows no new issues, PRs or branches in the scanned repos.
- [ ] 5.6 (user) Test the failure path with a non-existent repo in `DEPENDENCY_REPORT_REPOSITORIES`, then revert. Verify that the `❌ Renovate falló` message arrives.
- [ ] 5.7 (workflows) Delete `ops-dependency-report.yml` and its README row and section. Verify with `actionlint` on the remaining workflows.

## 6. Validation

- [ ] 6.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. The task passes only when both commands succeed on the current code.
