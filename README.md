# gh-export-secrets

[![GitHub Release](https://img.shields.io/github/v/release/katiem0/gh-export-secrets?style=flat&logo=github)](https://github.com/katiem0/gh-export-secrets/releases)
[![PR Checks](https://github.com/katiem0/gh-export-secrets/actions/workflows/main.yml/badge.svg)](https://github.com/katiem0/gh-export-secrets/actions/workflows/main.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/katiem0/gh-export-secrets)](https://goreportcard.com/report/github.com/katiem0/gh-export-secrets)
[![Go Version](https://img.shields.io/github/go-mod/go-version/katiem0/gh-export-secrets)](https://go.dev/)

A GitHub `gh` [CLI](https://cli.github.com/) extension to list the name and access level of GitHub
Actions, Dependabot, and Codepsaces secrets at the Organization and/or Repository level.

It produces a `csv` report detailing:

- `SecretLevel`: If the secret was created at the organization or repository level
- `SecretType`: If the secret was created for `Actions`, `Dependabot` or `Codespaces`
- `SecretName`: The name of the secret
- `SecretAccess`: If an organization level secret, the visibility of the secret
  (i.e. `all`, `private`, or `scoped`)
- `RepositoryName`: The name of the repository that the secret can be accessed from
- `RepositoryID`: The `id` of the repository that the secret can be accessed from

> **Note**
> This extension does **NOT** retrieve the value of the secret.

## Installation

1. Install the `gh` CLI - see the [installation](https://github.com/cli/cli#installation) instructions.

2. Install the extension:

    ```sh
    gh extension install katiem0/gh-export-secrets
    ```

For more information: [`gh extension install`](https://cli.github.com/manual/gh_extension_install)

## Usage

This extension supports `GitHub.com` and GHES, through the use of `--hostname`.

```sh
 $ gh export-secrets -h

Generate a report of Actions, Dependabot, and Codespaces secrets for an organization and/or repositories.

Usage:
  gh export-secrets [flags] <organization> [repo ...] 

Flags:
  -a, --app string           List secrets for a specific application or all: {all|actions|codespaces|dependabot} (default "actions")
  -d, --debug                To debug logging
  -h, --help                 help for gh
      --hostname string      GitHub Enterprise Server hostname (default "github.com")
  -o, --output-file string   Name of file to write CSV report (default "report-20230405134752.csv")
  -t, --token string         GitHub Personal Access Token (default "gh auth token")
```
