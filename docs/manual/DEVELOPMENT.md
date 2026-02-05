# Development

## Building the Provider
Supported dev platforms: Linux, macOS (Intel).

1. Clone the repository

1. Make sure that `GOPATH` is set in the environment.
    - It should be fine to use the path of this repository.
    - If you use **Nix** and `GOPATH` is not set, it will automatically set it to the repository path.

1. Enter the repository directory

1. Install requirements

1. Build the provider:
   - If you use **Nix**, remember to enter the shell: `./nix.sh`
   - Run `task build`

After the build, the provider binary should be available as `$GOPATH/bin/terraform-provider-privx`

## Adding Go Modules

This provider uses [Go modules](https://go.dev/wiki/Modules). Please see the Go documentation for the most up to date information about using them.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
task dep -- github.com/author/dependency

# Under the hood this runs:
# go get github.com/author/dependency
# go mod tidy
# go mod vendor
```

Remember to commit the changes to `vendor/`, `go.mod` and `go.sum`.

## Terraform Demo Project 
A Terraform `demo/` project that uses the provider can be build. You can use it to create/destroy resources in a PrivX instance.

**Note:** Make sure you use an instance of PrivX that is only needed for testing.

### Usage instructions:

- Set up the following in PrivX
  - A role with permissions to create resources
  - Set up API client access under _Administration -> Deployment -> Integrate with PrivX Using API Clients_

- Copy `.env-example` to `.env` and fill in the API and OAuth IDs and secrets.

- If you use **Nix**, remember to enter the shell: `./nix.sh`

- Run `task build-demo` to populate `demo/provider.tf` with PrivX credentials (not to be committed to Git).

- Before using Terraform with the demo project, we must tell Terraform where it can find our provider. We do so by adding the following to `$HOME/.terraformrc`:

    ```
    provider_installation {
      dev_overrides {
        "privx" = "$GOPATH/bin/"
      }
    }
    ```


**Note:** At the time of writing, generation of the demo project is still an ongoing activity.

### Create Documentation

Running `task doc` will generate the following documentation:

**Terraform Example**

In `examples/` you can find examples of resources that can be created with the provider. The demo project is built based on these resources. Using a Go module we create formatted documentation at `docs/resources/` of the resources.

**NOTE:** This is currently disabled in [Taskfile.yml](../../Taskfile.yml) as examples are currently being updated (Remove this line when it is done).

**Design diagrams**

Mermaid diagrams in `docs/manual/diagrams` are converted to SVGs. These files are used in the `docs/manual/DESIGN.md` document.

`mermaid-cli` NPM (NodeJs) module is used for the conversion.


### Running Terraform

Enable Terraform debug logging, and it will write a log file into the demo directory

```
 export TF_LOG=TRACE
 export TF_LOG_PATH=terraform.log
 
 terraform apply
 
```

### Configuring the Acceptance Test Suite
The tests executed by `task test` are controlled by the file:
`scripts/acceptance-tests.txt`

*   **To skip tests:** Comment out the test name with `#`.
*   **To stop after a specific test:** Place the word `exit` on a new line. Only tests listed *above* `exit` will be executed.

### Debugging tf files

When tests run, the HCL configurations used for each step are persisted to `internal/provider/tf_files/`. If a test fails, you can inspect these files to see exactly what was sent to the Terraform CLI.


## CI pipeline / GitHub Actions Configuration

To enable automated acceptance testing in GitHub Actions, you must configure the following **Repository Secrets** (Settings > Secrets and variables > Actions):

1.  **Secret Names:** Ensure the secret names exactly match those expected by the workflow:
    *   `PRIVX_API_BASE_URL`
    *   `PRIVX_API_OAUTH_CLIENT_ID`
    *   `PRIVX_API_OAUTH_CLIENT_SECRET`
    *   `PRIVX_API_CLIENT_ID`
    *   `PRIVX_API_CLIENT_SECRET`

2.  **Environment Connectivity:** The GitHub runner must be able to reach your PrivX instance at the URL specified in `PRIVX_API_BASE_URL`.
    
3.  **Github workflow example:** The [acceptance-tests](../../.github/workflows/acceptance-tests.yml) workflow.
