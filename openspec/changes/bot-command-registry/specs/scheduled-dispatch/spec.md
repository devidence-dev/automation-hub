# Spec Delta

## MODIFIED Requirements

### Requirement: Exact run link
Every successful dispatch, whether scheduled or from Telegram, SHALL identify the run it created from GitHub's dispatch response. It SHALL NOT pick the run by searching recent runs. If GitHub answers without run details, the link SHALL fall back to the repository's Actions page under the configured `github.web_url` (default `https://github.com`), as `<web_url>/<owner>/<repo>/actions`.

#### Scenario: Run details returned
- **WHEN** GitHub answers the dispatch with 200 and `html_url` `https://github.com/devidence-dev/workflows/actions/runs/123456`
- **THEN** the Telegram reply and the log use that URL

#### Scenario: Overlapping dispatches
- **WHEN** a scheduled and a manual `run_check_updates` are dispatched within the same second
- **THEN** each one reports the URL of its own run

#### Scenario: No run details
- **WHEN** GitHub answers the dispatch with 204 No Content
- **THEN** the dispatch counts as successful and the link is `https://github.com/devidence-dev/workflows/actions`

#### Scenario: Not a scheduled day
- **WHEN** the clock reaches Tuesday 09:00 America/Guayaquil for the same entry
- **THEN** no dispatch is sent

#### Scenario: Custom GitHub web URL
- **WHEN** `github.web_url` is `https://github.example.com/` (trailing slash removed) and GitHub answers the dispatch with 204 for `devidence-dev/workflows`
- **THEN** the dispatch counts as successful and the link is `https://github.example.com/devidence-dev/workflows/actions`

#### Scenario: Invalid GitHub web URL
- **WHEN** `github.web_url` is `github.com` (no scheme)
- **THEN** config loading fails with an error that names `github.web_url`
