# SendKit Go SDK

## Project Overview

Go SDK for the SendKit email API. Uses `net/http`, zero external dependencies.

## Architecture

```
sendkit.go     # Client: holds API key, HTTP client, doRequest()
emails.go      # EmailsService (Send, SendMime)
errors.go      # APIError type
```

- `NewClient()` creates client with API key + functional options
- `Client.Emails` exposes email operations
- All methods take `context.Context` as first parameter
- JSON struct tags use snake_case matching the API
- Errors returned as `*APIError` (implements `error` interface)
- `POST /v1/emails` for structured emails, `POST /v1/emails/mime` for raw MIME

## Testing

- Tests use `net/http/httptest` for mock HTTP servers
- Run tests: `go test ./...`
- No external test dependencies

## Releasing

- Tags use numeric format: `1.0.0` (no `v` prefix)
- CI runs tests on Go 1.21 and 1.23
- Pushing a tag creates a GitHub Release

## Git

- NEVER add `Co-Authored-By` lines to commit messages
