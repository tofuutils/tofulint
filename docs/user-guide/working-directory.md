# Switching working directory

TofuLint has `--chdir` and `--recursive` flags to inspect modules that are different from the current directory.

The `--chdir` flag is available just like OpenTofu:

```console
$ tofulint --chdir=environments/production
```

Its behavior is the same as [OpenTofu's behavior](https://developer.hashicorp.com/terraform/cli/commands#switching-working-directory-with-chdir). You should be aware of the following points:

- Config files are loaded after acting on the `--chdir` option.
  - This means that `tofulint --chdir=dir` will loads `dir/.tofulint.hcl` instead of `./.tofulint.hcl`.
- Relative paths are always resolved against the changed directory.
  - If you want to refer to the file in the original working directory, it is recommended to pass the absolute path using realpath(1) etc. e.g. `tofulint --config=$(realpath .tofulint.hcl)`.
- The `path.cwd` represents the original working directory. This is the same behavior as using `--chdir` in OpenTofu.

The `--recursive` flag enables recursive inspection. This is the same as running with `--chdir` for each directory.

```console
$ tofulint --recursive
```

These flags are also valid for `--init` and `--version`. Recursive init is required when installing required plugins all at once:

```console
$ tofulint --recursive --init
$ tofulint --recursive --version
$ tofulint --recursive
```
