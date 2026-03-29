# ferret

> **Work in progress.** ferret is under active development. Commands, keybindings, and file formats may change between versions.

Your API requests belong in your repo.

ferret is a terminal API client where collections are plain YAML files—sitting next to your code, versioned in git, and reviewable in a PR. No cloud accounts, no sync buttons, and no proprietary formats.

Open the TUI to explore interactively. Run requests in CI with the CLI. Same files, same environments, same behavior everywhere.

---

## Why ferret

Most API clients are heavy desktop IDEs. They're slow to open and tied to proprietary ecosystems. Your request history lives on their servers, not yours.

ferret is built for the terminal workflow. It opens instantly. It reads and writes plain YAML that you already know how to diff, grep, and version. When a teammate changes a request, you see it in the PR. When you need to run a request in CI, you use the same file you tested with locally.

No heavy runtimes. No gigabytes of overhead. Just a single binary that works over SSH, in a container, or on your local machine.

---

## Features

### Near-Zero Overhead
ferret is a single static binary with no cloud account, background sync process, or heavy runtime. It starts fast, runs anywhere your terminal does, and stays out of the way while you work.

### Git-native collections
Collections are plain YAML files in your repo, not a proprietary database. Review request changes with `git diff`, discuss them in pull requests, and version them with the rest of your code.

### Multi-tab request management
Keep multiple requests open side-by-side and switch instantly with `ctrl+n` / `ctrl+p`. Each tab keeps its own request state, so you can compare flows or iterate without losing in-progress work.

### Multi-collection workspaces
Point ferret at a parent directory and it discovers collections automatically. Jump between workspaces with `c` or open the collection picker with `C` to move across services quickly.

### CLI for automation
`ferret run` executes requests from collection files and exits cleanly for scripting. Use it in CI, pipe output into tools like `jq`, and reuse the exact same request definitions from local development.

### Keyboard-driven
Every primary workflow is accessible from the keyboard. Navigate panes, tabs, collections, and request fields without reaching for a mouse, with consistent terminal-first keybindings.

---

## Quick Start

```sh
# Install ferret using go (homebrew coming soon)
go install github.com/jxdones/ferret@latest

# Open ferret
ferret

# Execute a request via CLI
ferret run ./collections/login.yaml -e prod --raw | jq .
```

---

## Collection Layout

```text
my-collection/
  .ferret.yaml
  environments/
    dev.yaml
    prod.yaml
  requests/
    users/
      list.yaml
      create.yaml
```

Request file:

```yaml
name: list users
method: GET
url: "{{base_url}}/users"
```

Environment file:

```yaml
base_url: https://example.com
```

---

## Environments

- **TUI** (`ferret`): pass `--env` / `-e` to load `environments/<name>.yaml` from the active collection. Omit it and ferret loads the first env alphabetically if any exist; otherwise runs shell-only (OS env vars, no YAML layer).
- **`ferret run`**: pass `-e` with the env name. No auto-pick; empty name is not supported for file-backed envs.

---

## Multi-Collection Workspace

Point ferret at a parent directory containing multiple collections, then:

| Key | Action |
|-----|--------|
| `c` | cycle collections |
| `C` | open collection picker |
| `e` | cycle environments |
| `/` | search requests in active collection |

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `ctrl+r` | send request |
| `n` | new request |
| `/` | open request finder |
| `c` | cycle collection |
| `C` | pick collection |
| `e` | cycle environment |
| `m` | cycle HTTP method |
| `M` | open method picker |
| `ctrl+u` | focus URL bar |
| `tab` / `shift+tab` | cycle focus |
| `?` | full help |
| `q` / `ctrl+c` | quit |

### Tabs

| Key | Action |
|-----|--------|
| `ctrl+n` | next request tab |
| `ctrl+p` | previous request tab |
| `T` | new request tab |
| `X` | close request tab |

### URL bar

| Key | Action |
|-----|--------|
| `ctrl+l` | clear URL |
| `enter` / `esc` | leave URL bar |

### Panes

| Key | Action |
|-----|--------|
| `]` / `[` | next / previous pane tab |
| `j` / `k` | scroll |
| `ctrl+d` | scroll half page |
| `g` / `G` | top / bottom |

---

## Notes

- Default request headers are applied at send time when not set:
  - `Accept: */*`
  - `User-Agent: ferret`
- URLs without a scheme default to `https://`.
- Auth flows are currently manual — set the `Authorization` header using env vars.
