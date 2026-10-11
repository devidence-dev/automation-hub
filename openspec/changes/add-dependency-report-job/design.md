# Design

## Context

- Builds on `add-workflow-scheduler`: the dispatch core, the scheduler and `schedule` validation already exist.
- `main.go` builds the bot only when `workflows` or `restarts` is configured, and creates `kube.NewInClusterClient` only when `restarts` is. Webhook routes are registered through `switch hook.Name`, which only knows `qbittorrent`. The HTTP server has `ReadTimeout` and `WriteTimeout` of 15 s.
- `telegram.Client.SendMessage` hardcodes `ParseMode = "Markdown"` and retries up to 3 times.
- `internal/services/kube` is a minimal REST client (bearer token + CA from the ServiceAccount), with no client-go.
- `internal-apps` has a default-deny NetworkPolicy. Job → hub:8080 needs explicit egress and ingress rules (homelab).
- The interim jq in the `workflows` repo (`ops-dependency-report.yml`, uncommitted) is the reference for the message format.

## Goals / Non-Goals

**Goals:**
- The report format matches the interim jq byte for byte where possible, so the switch to the hub is invisible in Telegram.
- No new Go dependency.

**Non-Goals:**
- Generic Job orchestration (watching Job status, collecting logs). The Job reports back by itself through the webhook.

## Decisions

- **Job from a CronJob template with `suspend: true`.** The template lives in homelab, next to its RBAC and NetworkPolicy, and the hub only names it. Alternative: build the pod spec in the hub. It was rejected because it moves deployment details into app config and widens what the hub can create.
- **`CreateJobFromCronJob(ctx, ns, name) (string, error)` in `internal/services/kube`.** It sends a GET to `/apis/batch/v1/namespaces/{ns}/cronjobs/{name}` and a POST to `/apis/batch/v1/namespaces/{ns}/jobs` with `spec.jobTemplate` copied over. It mirrors what `kubectl create job --from` does. Verify the `cronjob.kubernetes.io/instantiate: manual` annotation in kubectl's `pkg/cmd/create/create_job.go` before using it. An ownerReference to the CronJob is optional; confirm what kubectl does and do the same.
- **Generalize the dispatch core into a per-command runner.** The bot and the scheduler look up a command and call its runner (workflow dispatch or job creation). The Telegram reply formatting stays per type.
- **Webhook answers 202 and sends in a goroutine.** N messages, each with up to 3 retries, can exceed the 15 s `WriteTimeout`, and the Job has `backoffLimit: 0`, so a synchronous 500 adds nothing. Alternative: raise the server timeout. It was rejected because it affects all routes and still has an upper bound. The goroutine uses a context with its own timeout, not the request context.
- **HTML parse mode through a new `SendHTML(chatID, text)`.** It shares the retry loop with `SendMessage` through a private helper. Existing Markdown callers stay as they are.
- **Formatting is pure functions.** `parseReport` and `buildMessages(report, forks) []string` take decoded input and return the message strings, so tests cover the format without HTTP or Telegram. Truncation works on unescaped lines and checks the escaped length per line, so a cut never lands inside an entity or tag.
- **Sender: the notifications bot (`telegram.bot_token`), not the workflow bot.** `dependency_report.telegram_chat_id` follows the same pattern as `hook[].config.telegram_chat_id`. Job-command replies keep using the workflow bot, like the other commands.
- **Route registration outside `switch hook.Name`.** `dependency_report:` is a top-level section with a fixed path, because its token and chat are not part of `WebhookProcessorConfig`.
- **Constant-time token check with `crypto/subtle`.** Body size is capped with `http.MaxBytesReader` (proposed: 10 MiB; a Renovate report for about 10 repos is well under that).

## Risks / Trade-offs

- [Renovate `reportType`/`reportPath` are experimental and the shape may change] → Pin Renovate to `44.151.0` in homelab, keep a synthetic report fixture in tests, and refresh the fixture from a real report when Renovate is upgraded.
- [`jobs create` RBAC cannot be limited by name] → Accepted for a single-user homelab, and documented in the ChangePlan. A ValidatingAdmissionPolicy could close it later.
- [Async send hides Telegram failures from the Job] → They are logged as errors in the hub. Accepted.
- [Webhook token is duplicated in hub config and the Job's ExternalSecret] → Rotating it means updating both. Documented.
- [Missing NetworkPolicy shows up only as a curl timeout in the reporter container] → Mandatory homelab task, checked on the first run.

## Migration Plan

1. homelab: ExternalSecret, CronJob (`suspend: true`), Role and RoleBinding, two NetworkPolicies.
2. Deploy the hub with `dependency_report:` and the `jobs:` entry in `CONFIG_YAML`.
3. Run `/run_dependency_report` manually, then delete the interim workflow from the `workflows` repo.
4. Rollback: remove the `jobs:` entry and the `dependency_report:` section, then delete the CronJob and ExternalSecret. No scanned repo is ever modified.

## Open Questions

- Should the report's root-level `problems` (not tied to any repo) be sent as a separate message? The interim jq ignores them. Default: ignore them, like the jq.
- Confirm that the target chat is the same one the PoC used. It only affects the config value.
