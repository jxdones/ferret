# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.2] - 2026-03-31

### Added

- **Per-tab collections.** Each tab now tracks its own active collection independently. Pressing `c` (cycle) or `C` (picker) changes only the active tab's collection, leaving other tabs unaffected. Opening a new tab starts with no collection selected. The first `c` press picks the first discovered collection.

### Changed

- **No collection pre-selected on startup.** Ferret no longer auto-selects the first collection when launching into a workspace. The title bar shows an empty collection label until the user selects one via `c` or `C`.

- **Workspace-level environments.** Pressing `e` now merges `environments/<name>.yaml` from every collection in the workspace into a single env. Variables from collections listed earlier in the workspace take precedence on key collisions; when a collision is detected the status bar shows `env -> <name> (key collisions)` as a hint. Switching a tab's collection no longer reloads the env.

## [0.2.1] - 2026-03-30

- **Per-tab loading state and request cancellation.** Each tab now tracks its own loading state independently. While a request is in flight, the status bar shows a spinner on the left and `^x to cancel` on the right. Pressing `ctrl+x` cancels the active tab's request immediately. Closing a tab with an in-flight request also cancels it automatically. Switching between tabs correctly reflects each tab's state. A loading tab resumes its spinner, a finished tab restores its response metadata.

### Fixed

- **Per-tab response isolation.** Responses now always land in the tab that issued the request, regardless of which tab is active when the response arrives. Previously, switching tabs while a request was in flight would cause the response to overwrite the wrong tab and steal focus. Each in-flight request now carries a stable tab ID so concurrent requests across multiple tabs resolve independently.

- **URL bar `enter` focus.** Pressing `enter` in the URL bar now moves focus to the request pane instead of the last active pane. Previously, if the response pane had been active, `enter` would send focus there instead of the request body editor.

## [0.2.0] - 2026-03-30

### Added

- **Large response protection.** Responses exceeding 10MB are no longer buffered into memory. The response pane shows the actual response size and a warning instead of rendering the body, preventing memory exhaustion on large or misconfigured API responses.

### Changed

- **HTTP request timeout.** Requests now have a 30-second default timeout. If the caller provides a context with its own deadline, that takes precedence. Previously, requests with no server response would hang indefinitely, freezing the TUI.
- **`exec.Execute` now accepts a `context.Context`.** The context is threaded through to the underlying HTTP request, enabling future per-tab cancellation support.

## [0.1.0] - 2026-03-29

### Added

- **Multi-tab request management.** Keep multiple requests open simultaneously. Each tab holds its own URL, method, headers, and response state independently. Switch tabs with `ctrl+n` / `ctrl+p`, open a new tab with `T`, and close with `X`. Tab labels show the HTTP method in its themed color and the request name or URL.

- **Request pane tabs.** The request pane exposes four tabs — `headers`, `params`, `body`, and `auth` — navigable with `]` / `[`. The params tab parses query parameters from the URL in real time.

- **Response pane tabs.** The response pane exposes four tabs — `body`, `headers`, `cookies`, and `trace`. The body tab renders JSON, XML, and HTML with syntax highlighting. The trace tab shows a per-stage timing breakdown and redirect history.

- **Response syntax highlighting.** Response bodies are syntax-highlighted using [chroma](https://github.com/alecthomas/chroma). The lexer is auto-detected from the `Content-Type` header, falling back to content sniffing.

- **Environment variable interpolation.** Use `{{variable}}` placeholders in request URLs, headers, and bodies. Values are resolved from a layered environment: YAML file, session variables, and OS shell env vars.

- **Multi-collection workspace.** Point ferret at a parent directory and it discovers all collections underneath. Cycle with `c` or open the picker with `C`.

- **Environment cycling.** Press `e` to cycle through environments defined in `environments/*.yaml`. Switching environments preserves session variables extracted from previous responses.

- **Method picker modal.** Press `M` to open a modal and select the HTTP method. Press `m` to cycle through `GET → POST → PUT → PATCH → DELETE`.

- **URL bar with paste support.** The URL bar accepts typed input, paste via `ctrl+v`, and can be cleared with `ctrl+l`.

- **`https://` fallback.** URLs without a scheme automatically get `https://` prepended at send time.

- **Request trace.** Every request records a per-stage timing timeline (DNS, connect, TLS handshake, TTFB, body read) visible in the response trace tab.

- **Status bar.** Shows request status, HTTP status code, response size, and duration after each send. Displays warnings and errors inline.

- **`ferret run` CLI.** Run a single request from a collection file and print the response to stdout. Supports `--raw` for piping to `jq` and `-e` for environment selection.

- **Keyboard-driven navigation.** Full keyboard control throughout. `tab` / `shift+tab` cycles focus between the URL bar, request pane, and response pane. `?` opens the full help overlay.
