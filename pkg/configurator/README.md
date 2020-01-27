# KubeKit Configurator Go Package

This is a [KubeKit](https://kubekit.dev/)
Go package to bring up Kubernetes on a given set of infrastructure. It is used by the KubeKit CLI and the KubeKit Configurator Microservice.

The Go package encapsulate Ansible code which is executed on every Kubernetes node.

## Requirements

* [Go](https://golang.org/doc/install) or [Docker](https://docs.docker.com/install/) are required to verify the Go code compile
* (Optional) [Ansible](https://docs.ansible.com/ansible-container/installation.html) [will be] required to execute Unit Tests to the Ansible code.

This repository/project does not store the vendors or imported Go packages, it will require all the vendors in your system if you would like to compile it using the Go in your system. If Go is not installed or you don't have the vendors in your system, the **recomended** way to verify the code compile is using the Docker container.

## Build

If the changes were done to the Ansible code located in the `templates/` directory, the Go code that encapsulate such code has to be generated. Execute `make generate` or just `make`:

    make generate

If the changes were done to the Go code there are two ways to compile:

1. If you have **Go install in your system and the vendors correctly setup**: Execute `make install` or `make test` or both `make test install`.
2. If Go is not installed but Docker is: Execute `make compile`

The safest way to proceed is the option #2 but takes more time (2-3 minutes):

    make compile

It's recommended to use an IDE such as VSCode or Atom with the Golang plugings to have a well formatted code, otherwise, execute `make fmt` before compile and push the code to GitHub.

## Use

The configurator has a Config struct with all the required parameters to configure Kubernetes in a node or cluster. That Config struct is created with the method `New()` giving the parameters in a `map[string]interface{}` map.

To execute the configurator, call the method `Configure()` to install Ansible on every node (if not already installed) and to send the certificates, manifest and Ansible roles to every Kubernetes node. Then it will execute the Ansible playbook to finally install, configure and start up Kubernetes in every node.

    import (
      "github.com/liferaft/configurator"
    )

    func Configure() error {
      logger.Debugf("starting the configuration of cluster %q", Name)

      configtor, err := configurator.New(Config, state, path, Platform, logger)
      if err != nil {
        return err
      }

      return configtor.Configure()
    }

## Setup Vendors

**IMPORTAT**: The execution of these actions may modify the existing Go packages you have in your $GOPATH directory. **Use cafefully**.

If you don't want to modify the existing vendors, use the compilation in a container. It's safer but takes longer (2-3 minutes).

To quickly and easly setup the required Go packages in your system, execute `make vendor`. This will get or update the required Go packages (vendors) in your $GOPATH directory. Then will fix some errors that cause the code don't compile.

## Go Vendor Problems

The vendor actions, like getting new vendors or updates, sometimes may cause compilation errors. Some possible causes are:

* Repeated packages: Some packages are inside the vendor directory of a package and in the $GOPATH. To remove the package from the vendor directory may fix this issue.
* Not vendored packages: Some packages are not added correctly to the project vendor directory by Govendor. If this is the case, copy the package from $GOPATH but make sure to not copy the .git directory and files.
* Sirupsen: The `Sirupsen` logging package change to `sirupsen` but some packages still uses the former. Change the import to use `sirupsen`.
