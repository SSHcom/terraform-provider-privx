# Requirements 
### Development Dependencies installed manually

- [Nix](https://nixos.org/download/) = version >= 2.15.0
- [Taskfile](https://taskfile.dev/docs/installation) = latest (sets up runnable tasks in Taskfile.yml)
- [Godotenv](https://github.com/joho/godotenv) = latest

### Development Dependencies installed by Nix
You can use [Nix](https://github.com/NixOS/nix) (version >= 2.15.0) to easily provide all dependencies in a reproducible development environment.
**Note:** It is recommended to do the single-user installation. If you decide on the multi-user option, be aware it may not work if SELinux is enabled.

Running `./nix.sh` enters a shell session where the requirements listed below will be available

- [Terraform](https://www.terraform.io/downloads.html) >= 1.13
- [Go](https://golang.org/doc/install) >= 1.25
- [Golangci-lint](https://github.com/golangci/golangci-lint) >= 1.53.x
- [Pre-commit](https://github.com/pre-commit/pre-commit) >= 3.3.x
- [NodeJs](https://nodejs.org/en) >= 20 (needed for mermaid-cli)
- [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli) = latest (Used for converting Mermaid diagram ".mmd" files to SVG)


#### Environment Variables
- When entering the Nix shell, your existing environment variables are preserved.
- `$GOPATH` is set to the repository directory if empty
- `PATH` is modified to include `$GOPATH/bin`
- To include additional environment variables for the Nix shell, add them to the `shellHook` section of the `shell.nix` file.

#### Using direnv (optional)

- If you use [direnv](https://direnv.net/), you can create a `.envrc` file with `use_nix` to automatically load the Nix shell when entering the directory. See `envrc.EXAMPLE`.