# Contributing Guide

## Go Code Contributions

Refer to the [Go Best Practices guide](https://github.com/kraken/goblets) for anything not covered in this contributing guide.

### [WIP] Dependency Management

**NOTE: the following section is a WIP that will not be in affect until after UKS-1423. This notice along with the `[WIP]` tag should be removed when UKS-1423 is complete.**

KubeKit utilizes the versioned Go module system (modules) to maintain versions of its dependencies. Go modules require Go v1.11.

Refer to the [Go modules best practices guide](https://github.com/kraken/goblets/blob/master/docs/modules.md) for more detailed information pertaining to modules.

#### Updating dependencies

When making changes to the code, the dependency graph may change and require updates to the mod file (`go.mod`) and vendor folder. When this occurs, run the following makefile target:

`make vendor-mod`

**NOTE:** the vendor folder should **NEVER** be modified directly. It should be generated as an artifact of the above command to maintain consistency with the mod file.

#### Local Workflows

When a developer needs to update dependencies in KubeKit to use a local fork for development purposes, the following workflows can be followed to allow Go modules to work.

| Local fork | Set command | Unset Command |
|------------|------------------|-----------------|
| github.com/kubekit/provisioner | `make local-provisioner-on` | `make local-provisioner-off` |
| github.com/kubekit/configurator | `make local-configurator-on` | `make local-configurator-off` |
| github.com/kubekit/manifest | `make local-manifest-on` | `make local-manifest-off` |
| github.com/kraken/terraformer | `make local-terraformer-on` | `make local-terraformer-off` |
| All of the above | `make local-all-on` | `make local-all-off` |

The "set" command will enable a local workflow. It does this by first backing up the `go.mod` file and then creating the necessary change to the working copy of `go.mod`. When finished working with the local fork, **always** revert the change with the "unset" command.

**IMPORTANT:** Do not commit changes to `go.mod` or `vendor` while utilizing local workflows.
