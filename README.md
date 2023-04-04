# gh-export-secrets

A GitHub `gh` [CLI](https://cli.github.com/) extension to list the names and access level for GitHub Actions, Dependabot, and Codepsaces secrets at the Organization and/or Repository level.


It produces a `csv` report detailing:

- `SecretLevel`: If the secret was created at the organization or repository level
- `SecretType`: If the secret was created for `Actions`, `Dependabot` or `Codespaces`
- `SecretName`: The name of the secret
- `SecretAccess`: If an organization level secret, the visibility of the secret (i.e. `all`, `private`, or `scoped`)
- `RepositoryName`: The name of the repository that the secret can be accessed from 
- `RepositoryID`: The `id` of the repository that the secret can be accessed from 

> **Note:**
> This extension does **NOT** retrieve the value of the secret.

## Installation

1. Install the `gh` CLI - see the [installation](https://github.com/cli/cli#installation) instructions.

2. Install the extension:

    ```sh
    gh extension install katiem0/gh-export-secrets
    ```

For more information: [`gh extension install`](https://cli.github.com/manual/gh_extension_install)

## Usage

> **Note:** 
> This extension only supports github.com.

```sh
 $ gh export-secrets -h

Generate a report of Actions, Dependabot, and Codespaces secrets for an organization and/or repositories.

Usage:
  gh export-secrets [flags] <organization> [repo ...] 

Flags:
  -b, --actionsSecrets       To retrieve Actions secrets
  -a, --all                  To retrieve all secrets types
  -c, --codespacesSecrets    To retrieve Codespaces secrets
  -e, --debug                To debug logging
  -d, --dependabotSecrets    To retrieve Dependabot secrets
  -h, --help                 help for gh
  -o, --output-file string   Name of file to write CSV report (default "report-20230404120355.csv")
```
