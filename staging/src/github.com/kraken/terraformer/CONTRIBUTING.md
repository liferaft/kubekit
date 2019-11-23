# Contributing Guide

## Dependency Management

This project uses Go modules (a.k.a. "vgo") to define all dependent libraries. To use modules, you will need Go v1.11 or higher, or the [vgo command line tool](https://github.com/golang/vgo/). Also, the environment variable `GO111MODULE` must be set to `on` to indicate that a project should be built using modules:

`export GO111MODULE=on`

All dependencies are captured in the [go.mod](go.mod) and [go.sum](go.sum) files. When making updates to the dependencies, these two files should be updated in lieu of vendoring dependencies or using another package manager (e.g. govendor).

At the time of this writing (2018-08-30) there is lacking 3rd party Go developer tool support for vgo. Your IDE may fail to recognize the correct location of imported dependencies. As a work around, you can vendor your dependencies the old fashion way into your project space with this Makefile command:

`make vendor-vgo`

Refer to the Kraken team's [vgo best practices](https://github.com/kraken/goblets/blob/master/docs/vgo.md) for more details.

### Terraform Versioning Caveats

Terraformer depends on [the Hashicorp library Terraform](https://github.com/hashicorp/terraform). Teraform allows for different "providers" to allow teraforming capabilities for different platforms. Each of these providers is hosted separately in different repos with different dependent versions of Terraform. This causes incompatability issues when using automated dependency management.

Most dependency managers create a tree structure of all dependencies and then flatten the tree to only import one version of each dependency. If there are conflicts in the versions required by a single project, the manager may fail to resolve dependencies or there may be build errors.