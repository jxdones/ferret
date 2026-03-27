# ferret

Your API requests belong in your repo.

ferret is a terminal API client where collections are just YAML files on disk — sitting next to your code, versioned in git, readable in a diff, reviewable in a PR. No proprietary format, no sync button, no cloud account required.

Open the TUI to explore interactively. Use the CLI to run requests in CI. Same files, same environments, same behavior.

One binary. Drop it in your path and it works.

## Collection Layout

Each collection directory should look like:

```text
my-collection/
  .ferret.yaml
  environments/
    dev.yaml
  requests/
    users/
      list.yaml
```

Request file example:

```yaml
name: get ditto
method: GET
url: "{{poke_base_url}}/pokemon/ditto"
```

Environment file example:

```yaml
poke_base_url: https://pokeapi.co/api/v2
```

