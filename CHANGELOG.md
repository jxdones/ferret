# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.8] - 2026-04-23

### Added

- **Harbor theme.** A new built-in theme with a dark-navy canvas, vivid teal accents, hot-pink POST highlights, and sharp green GET/success colors. Set `theme: harbor` in `~/.config/ferret/config.yaml` to enable it. Harbor is the first theme to use the new `Background` field, which sets the terminal's background color via `tea.View.BackgroundColor` so the canvas color is consistent regardless of the terminal's own default.

- **Theme background color support.** The `Theme` struct now has an optional `Background` field. When set, ferret passes it to bubbletea's `View.BackgroundColor`, painting the terminal's actual background rather than relying on the terminal default. Existing themes leave `Background` nil and are unaffected.

### Changed

- **Context-sensitive shortcuts bar for the headers tab.** The inline hint (`i/I/A add · d delete row · tab/shift+tab fields`) has been removed from the headers content area. The bottom shortcuts bar now shows the relevant bindings dynamically: in normal mode it shows `]/[`, `j/k navigate`, `i/I/A add row`, and `d delete row`. In insert mode it switches to `tab next field`, `shift+tab prev field`, `enter commit`, and `esc cancel`.

## [0.2.7] - 2026-04-12

### Added

- **Closing the last tab opens a fresh one.** Pressing the close-tab key when only one tab is open no longer blocks. The existing tab is replaced with a brand-new empty request tab, the same state as when ferret first launches.

- **Auth tab in the request pane.** The `auth` tab now renders a live editor for the active request's auth configuration. Supports `bearer`, `basic`, `api key` (header or query), and `none`. The auth type is selected with `h`/`l`, fields are navigated with `j`/`k`, and edited with `i` or `enter`. Loading a request from the collection picker populates the auth tab from the request's `auth` field, inheriting from `.ferret.yaml` when the request does not override it.

### Changed

- **Sanitized filesystem paths in error messages.** Errors from `collection` and `env` no longer expose full absolute paths (which leak usernames and directory layout). File errors now show a short `collection/file.yaml` style identifier: direct children of `requests/` or `environments/` use the collection name as prefix (e.g. `pokeapi/list.yaml`), nested files use their immediate parent directory (e.g. `users/list.yaml`). Directory-level errors use generic messages.

## [0.2.6] - 2026-04-07

### Fixed

- **Wide-character cursor misalignment in the body editor.** The cursor highlight was positioned by rune count instead of display columns, causing it to land in the wrong cell when the body contained CJK, hangul characters or emoji (each 2 display columns wide). The render loop now advances by `uniseg.StringWidth` per rune, matching the display-column offset reported by the textarea.

### Added

- **Pre-request hooks.** Requests and collections now support a `pre_request` field pointing to an executable script. Before each request is sent, ferret runs the script and merges any `KEY=value` lines from stdout into the session environment, making those variables available for interpolation in the URL, headers, body, and auth fields. Scripts are shebang-based and language-agnostic (shell, Python, Node, etc.). The full ferret environment (file, session, and shell layers) is passed to the script as process environment variables so hooks can reference `{{variables}}` directly. Hook resolution follows the same inherit/override pattern as auth: omitting `pre_request` on a request means no hook runs. Setting it to `inherit` uses the collection default from `.ferret.yaml`. Any path runs that specific script.

- **Hook status in the TUI.** While a hook is running the status bar shows `Running <script-name>` in a distinct color with a spinner. Once the hook finishes the status transitions to the normal `Making request` state. Pressing `^x` cancels the hook mid-run. Hooks time out after 10 seconds.

## [0.2.5] - 2026-04-07

- **Add theme support.** Now themes are set on the config file. We added 6 new themes to ferret `princess`, `dracula`, `catppuccin`, `gruvbox`, `solarized`, `everforest`.

### Added

- **First-class auth support.** Requests and collections now support a structured `auth` field with four built-in types: `bearer`, `basic`, `apikey` (header or query param), and `none`. Auth values support `{{variable}}` interpolation. Request-level auth overrides the collection default; omitting `auth` on a request inherits from `.ferret.yaml`. The old `auth: string` stub on `Request` and `AuthConfig` on `Config` have been replaced with a unified `Auth` struct.

## [0.2.4] - 2026-04-05

### Added

- **Horizontal scroll in the response body tab.** The body pane now supports horizontal scrolling via `h`/`l` (or left/right arrows). `0` jumps to the beginning of the line and `$` jumps to the end of the longest line. The horizontal offset resets automatically when a new response arrives.

### Changed

- **`LoadConfig` no longer writes on read.** `LoadConfig` now returns `DefaultConfig()` when the config file is missing instead of creating it on disk. A new `EnsureConfigExists` function owns the creation logic and is called explicitly at startup. This makes `LoadConfig` safe to call in any context without unexpected filesystem side effects.

- **`ListNames` now returns a sorted slice.** `env.ListNames` previously documented undefined ordering and required every caller to sort. The function now sorts internally and callers no longer need to.

- **Unified `TruncatePad` utility.** Five independent "truncate-or-pad to N columns" implementations (`urlbar.fit`, `headers.padRight`, `requestpane.padRight`, `view.fitStyled`, `view.fitToWidth`) have been replaced with a single `common.TruncatePad`.

- **Unified `DetectSyntax` utility.** `requestpane` and `responsepane` each had ~40 lines of duplicated content-type heuristics. Both now delegate to a single `common.DetectSyntax(contentType, body string) string` that checks the Content-Type header first, then falls back to body sniffing. As a side effect, the request pane now also recognises XML and HTML bodies when no Content-Type header is set.

- **Unified `FormatSize` utility.** `statusbar` and `responsepane` each had their own `formatSize` — the statusbar version was missing the MB branch (a 10 MB response would render as `10485.8KB`). Both have been replaced with a single `common.FormatSize` that correctly handles B / KB / MB.

### Fixed

- **Collection modal shows stale entries on collection-agnostic tabs.** Pressing `/` on a tab with no collection loaded previously displayed entries from another tab's collection. The `collection.Model` is shared across tabs, and `Reset()` cleared the search input but left `all` intact. The fix clears the entry list before attempting to load for the active tab, so tabs with no collection root always open an empty modal.

- **Tab title panic on multibyte URLs.** `clampTabTitle` previously sliced the URL string at byte offset 10, which panics when a multibyte character (emoji, CJK, accented) straddles that boundary. The function now uses `go-runewidth` to measure and truncate by display columns, handling all Unicode correctly. The `https://` and `http://` scheme prefix is also stripped before clamping so the visible columns are spent on the meaningful part of the URL.

## [0.2.3] - 2026-04-02

### Changed

- **Context-aware shortcut hints.** The bottom shortcuts bar and the expanded `?` help view now lead with bindings for the focused area (URL bar, request pane, or response pane), then list global shortcuts. The compact bar keeps a small set of always-visible globals (`^r` send, `?` help) alongside pane-specific keys (for example `j`/`k` and `]/[` in the panes). Full help still documents tab focus, collections, tabs, and quit; `esc` (clear focus) and `q` (quit) appear in the global section.

### Fixed

- **`ferret run` requires `--env`/`-e`.** Omitting the flag previously produced a confusing error that leaked the collection's filesystem path. The flag is now marked required by Cobra, which surfaces a clean `required flag(s) "env" not set` error before any internal code runs.

### Security

- **Tighter file permissions for saved requests.** `SaveRequest` now creates directories with `0o700` (owner-only) and writes request files with `0o600` (owner-only read/write), down from `0o755`/`0o644`. Prevents other users on the same system from reading request files that may contain auth tokens or sensitive URLs.

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
