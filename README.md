# terraformat

A standalone Terraform formatter that goes beyond `terraform fmt`:

- Orders top-level blocks (terraform, provider, variable, locals, data, resource, module, output).
- Normalizes blank lines between top-level blocks.
- Reorders attributes inside common block types (resource, variable, output, module, provider, terraform, locals).
- Ensures a trailing newline at EOF.

## Install

```
go install github.com/mreimbold/terraformat/cmd/terraformat@latest
```

## Usage

```
terraformat -write=false path/to/file.tf
terraformat -check path/to/dir
cat file.tf | terraformat
```


## Notes

- Only `.tf` and `.tfvars` files are formatted.
- Directories are walked recursively, skipping hidden directories and `.terraform`.
