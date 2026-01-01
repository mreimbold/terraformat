# terraformat

[![CI](https://github.com/mreimbold/terraformat/actions/workflows/ci.yml/badge.svg)](https://github.com/mreimbold/terraformat/actions/workflows/ci.yml)
[![Release](https://github.com/mreimbold/terraformat/actions/workflows/release.yml/badge.svg)](https://github.com/mreimbold/terraformat/actions/workflows/release.yml)

terraformat is a standalone Terraform/OpenTofu formatter that goes beyond
`terraform fmt` (and `tofu fmt`). It enforces the parts of the official
[Terraform language style guide](https://developer.hashicorp.com/terraform/language/syntax/style)
the built-in formatter does not. The goal is fully
deterministic formatting that ends style debates in reviews.

It is a drop-in replacement for `terraform fmt` / `tofu fmt`.

## Features

Rules enforced here that the official formatters do not enforce:

- Orders top-level blocks (terraform, provider, variable, locals, data, resource,
  module, output).
- Normalizes blank lines between top-level blocks and logical sections.
- Orders attributes in common blocks (resource, variable, output, module,
  provider, terraform).
- Preserves comments and produces idempotent output.
- Ensures a trailing newline at EOF.

## Install

### Go install

```bash
go install github.com/mreimbold/terraformat/cmd/terraformat@latest
```

### Releases

Download a prebuilt binary from the GitHub Releases page.

## Usage

```bash
terraformat [options] [target...]
```

Targets behave like `terraform fmt`:

- No target: format the current directory.
- Directory target: format files in that directory.
- File target: format that file only.
- A single dash (`-`): read from STDIN.

Options (compatible with `terraform fmt`):

- `-list=false`    don't list files whose formatting differs (disabled on STDIN)
- `-write=false`   don't write to source files (disabled on STDIN or `-check`)
- `-diff`          display diffs of formatting changes
- `-check`         exit non-zero if any input is not formatted
- `-no-color`      disable colored output
- `-recursive`     process subdirectories (default: current directory only)

Exit codes:

- `0` success
- `1` flag/usage error
- `2` formatting error
- `3` files would change (`-check`)

## Examples

```bash
terraformat -write=false path/to/file.tf
terraformat -check -recursive path/to/module
terraformat -diff -write=false path/to/file.tf
cat file.tf | terraformat
```

## Supported files

- `.tf`, `.tfvars`, `.tftest.hcl`, `.tfmock.hcl`
- JSON files (`.tf.json`, `.tfvars.json`, `.tftest.json`) are not supported
- Hidden directories and `.terraform` are skipped

## Prerequisites

- Go 1.25+
- A `diff` command on PATH if you use `-diff`

## License

MIT
