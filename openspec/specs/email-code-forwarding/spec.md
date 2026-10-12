# email-code-forwarding Specification

## Purpose
Forwards one-time codes from unread emails to Telegram. The hub polls an IMAP inbox, matches each unread email to a configured service by sender and subject, extracts the code with that service's configured strategy, and sends the service's message. Documents the current behavior and the per-service settings that replace the built-in, name-based defaults.

## Requirements

### Requirement: Polling interval
The system SHALL poll the IMAP inbox every `email.polling_interval` seconds. When the key is absent or `0`, the interval MUST be 60 seconds. A negative value MUST fail config loading with an error that names `email.polling_interval`.

#### Scenario: Default interval
- **WHEN** the config has no `polling_interval` under `email:`
- **THEN** the first poll runs 60 seconds after startup and then every 60 seconds

#### Scenario: Explicit interval
- **WHEN** `email.polling_interval` is `20`
- **THEN** the inbox is polled every 20 seconds

#### Scenario: Negative interval
- **WHEN** `email.polling_interval` is `-5`
- **THEN** config loading fails with an error that names `email.polling_interval`

### Requirement: Unread email search
Each poll SHALL search the INBOX for unread emails whose From address contains each configured service's `email_from`, and process the union of the results. When no service is configured, the poll SHALL search all unread emails. A failed search for one sender MUST be logged with that sender, and the remaining senders MUST still be searched. A failed connection, login or INBOX select MUST be logged with the IMAP host and error, and the poll MUST end without processing anything. The next poll retries. Every email service MUST have a non-empty `email_from`, because an empty value would match every sender; config loading MUST fail otherwise, with an error that names the service.

#### Scenario: Two services, one unrelated email
- **WHEN** service `cloudflare` has `email_from: "noreply@notify.cloudflare.com"`, service `openai` has `email_from: "noreply@tm.openai.com"`, and the inbox holds one unread email from each sender plus one unread email from `news@example.org`
- **THEN** only the two matching emails are fetched and offered to the processors

#### Scenario: One sender search fails
- **WHEN** the search for `noreply@tm.openai.com` fails and the search for `noreply@notify.cloudflare.com` succeeds
- **THEN** the failed search is logged with `sender=noreply@tm.openai.com` and the Cloudflare email is still processed

#### Scenario: Login fails
- **WHEN** the IMAP server rejects the login for `imap.gmail.com` at 09:00:00 America/Guayaquil
- **THEN** an error naming host `imap.gmail.com` is logged, no email is processed, and the next poll at 09:00:20 retries

#### Scenario: Service without email_from
- **WHEN** service `github` has no `email_from` key, or an empty one
- **THEN** config loading fails with an error that names service `github`

### Requirement: Processor dispatch
Each fetched email SHALL be offered to the configured services in config order. The first service whose `email_from` is contained in the From address, and whose `email_subject` list contains a value contained in the Subject, handles the email. An email that no service handles MUST be left unread and MUST NOT be sent to Telegram.

#### Scenario: Cloudflare login code
- **WHEN** an unread email from `noreply@notify.cloudflare.com` with subject `Your login code for devidence.dev` arrives, and the body contains `Use this code directly: 123456`
- **THEN** the Telegram chat receives `🛡️ Cloudflare App Code: ` followed by a code block containing `123456`

#### Scenario: Unrelated sender
- **WHEN** an unread email from `other@domain.com` with subject `Your login code for devidence.dev` arrives
- **THEN** no service handles it, no Telegram message is sent, and the email stays unread

#### Scenario: Unmatched subject
- **WHEN** an unread email from `noreply@notify.cloudflare.com` with subject `Weekly newsletter` arrives
- **THEN** no service handles it, no Telegram message is sent, and the email stays unread

### Requirement: Body preparation
Before extraction, quoted-printable content SHALL be decoded. Unless the service sets `strip_mime_headers: false`, the text up to and including the first blank line (`\n\n` or `\r\n\r\n`) MUST be removed. `strip_mime_headers` defaults to `true`. If the text has no blank line, it is used unchanged.

#### Scenario: Quoted-printable body
- **WHEN** the body is `Hello=20World=21`
- **THEN** extraction runs on `Hello World!`

#### Scenario: Headers stripped by default
- **WHEN** a service without `strip_mime_headers` receives the text `From: foo\r\nSubject: bar\r\n\r\nBody content`
- **THEN** extraction runs on `Body content`

#### Scenario: Headers kept for Cloudflare
- **WHEN** service `cloudflare` sets `strip_mime_headers: false` and the text is `Subject: Code 111111\n\nYour code is 222222`
- **THEN** the `\b\d{6}\b` pattern extracts `111111`

### Requirement: Pattern extraction
A service with `extractor: pattern`, which is the default, SHALL extract the first match of `code_pattern` in the prepared text. When the pattern has a capture group that matched a non-empty value, that group is the code; otherwise the whole match is the code. When `code_pattern` is empty, the pattern MUST be `\b[a-zA-Z0-9]{4,8}\b`. An invalid `code_pattern` MUST fail config loading with an error that names the service.

#### Scenario: Hex color is not a code
- **WHEN** service `openai` has `code_pattern: "[^#]\\b(\\d{6})\\b"` and the text is `color:#202123; padding: 0\nIngresa este codigo:\n662575\n`
- **THEN** the extracted code is `662575`

#### Scenario: Default pattern
- **WHEN** service `github` has no `code_pattern` and the text is `Use 4821 now`
- **THEN** the extracted code is `4821`

#### Scenario: Invalid pattern
- **WHEN** service `github` has `code_pattern: "[invalid regex ("`
- **THEN** config loading fails with an error that names service `github`

### Requirement: Marker extraction
A service with `extractor: marker` SHALL find the first entry of `code_markers`, in configured order, that occurs in the prepared text, compared case-insensitively. It SHALL then try each entry of `code_candidates`, in order, on the text after that marker; the first candidate that matches gives the code from its first capture group. If no candidate matches, `code_pattern` (when set) is the last fallback. `extractor: marker` with an empty `code_markers` MUST fail config loading. An `extractor` value other than `pattern` or `marker` MUST fail config loading.

#### Scenario: Numeric code after a Spanish marker
- **WHEN** service `perplexity` has `code_markers: ["directly:", "directamente:"]` and the text is `Inicie sesión directamente:\n123456`
- **THEN** the extracted code is `123456`

#### Scenario: Hyphenated code after an English marker
- **WHEN** the text is `Log in directly:\naw9s5-y1zoy`
- **THEN** the extracted code is `aw9s5-y1zoy`

#### Scenario: Marker order wins over position
- **WHEN** the text is `directamente: 111111 ... directly: 222222`
- **THEN** the extracted code is `222222`, because `directly:` is first in `code_markers`

#### Scenario: Marker missing
- **WHEN** the text is `Welcome to Perplexity. Click here to confirm.`
- **THEN** no code is found

#### Scenario: Marker present without a valid code
- **WHEN** the text is `Log in directly: !!!`
- **THEN** no code is found

#### Scenario: Marker without markers
- **WHEN** service `perplexity` has `extractor: marker` and no `code_markers`
- **THEN** config loading fails with an error that names service `perplexity`

#### Scenario: Unknown extractor
- **WHEN** service `perplexity` has `extractor: regex`
- **THEN** config loading fails with an error that names the value `regex`

### Requirement: Not-found text
When no code is found, the service SHALL send its `not_found_text` through its `telegram_message` instead of a code. `not_found_text` defaults to `Not found`. The processor MUST log a warning with a preview of at most `email.log_preview_length` characters (default 200) of the text it searched. A negative `email.log_preview_length` MUST fail config loading with an error that names `email.log_preview_length`; zero or absent means 200. A not-found result is a successful processing run, so the mark-as-read policy applies.

#### Scenario: Configured not-found text
- **WHEN** service `openai` has `not_found_text: "No code in email"`, its template is `🤖 ChatGPT Code: ` followed by a code block holding `%s`, and the text contains no code
- **THEN** the chat receives the same template with `No code in email` in place of `%s`

#### Scenario: Default not-found text
- **WHEN** service `cloudflare` has no `not_found_text`, its template is the Cloudflare template from the config example, and the text contains no code
- **THEN** the chat receives that template with `Not found` in place of `%s`

### Requirement: Telegram notification
The service SHALL format the code into its `telegram_message`, which has one `%s` placeholder, and send it to `telegram_chat_id`. When the send fails, the error MUST be returned, the failure MUST be logged with the subject, and the email MUST NOT be marked as read.

#### Scenario: Successful forward
- **WHEN** service `openai` with `telegram_chat_id: "123456789"` and template `🤖 ChatGPT Code: ` followed by a code block holding `%s` extracts `662575`
- **THEN** chat `123456789` receives `🤖 ChatGPT Code:` followed by a code block holding `662575`

#### Scenario: Telegram rejects the message
- **WHEN** the Telegram API returns an error for the send
- **THEN** the failure is logged with the email subject, and the email stays unread so the next poll retries it

### Requirement: Mark-as-read policy
After a successful processing run, the matched service's `mark_as_read` decides the email's Seen flag. The flag MUST be set when `mark_as_read` is unset or `true`, and MUST NOT be set when it is `false`. A failed processing run MUST NOT set the flag.

#### Scenario: Unset means read
- **WHEN** the Cloudflare email is forwarded and `mark_as_read` is unset
- **THEN** the email gets the `\Seen` flag

#### Scenario: Explicit false
- **WHEN** service `openai` has `mark_as_read: false` and the forward succeeds
- **THEN** the email stays unread, and the next poll forwards it again

#### Scenario: Failed forward
- **WHEN** the Telegram send fails
- **THEN** the email stays unread

### Requirement: Malformed message handling
A fetched message without an envelope MUST be processed with an empty ID, subject and sender. It MUST NOT cause a panic. It matches no service that has a non-empty `email_from`, so it stays unread.

#### Scenario: Message without envelope
- **WHEN** the server returns a message with no envelope and body `123456`
- **THEN** the poll completes, no Telegram message is sent, and the message stays unread

### Requirement: Extractor configuration validation
Config loading MUST fail, with an error that names the service and key, when a service has an invalid regex in `code_pattern` or `code_candidates`, an unknown `extractor`, or `extractor: marker` without `code_markers`. A service with none of the optional keys MUST load with the defaults in the Pattern extraction and Not-found text requirements.

#### Scenario: Valid minimal service
- **WHEN** service `github` has only `email_from`, `email_subject`, `telegram_chat_id` and `telegram_message`
- **THEN** config loading succeeds and the service uses `extractor: pattern`, the default pattern and `Not found`
