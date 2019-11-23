# KubeKit Provisioner Go Package

This is a [KubeKit](https://kubekit.dev)
Go package to create a Kubernetes-ready infrastructure on different platforms or clouds. It is used by the KubeKit CLI and the KubeKit Provisioner Microservice.

The Go package encapsulate Terraform code which is executed on the user computer, however, **Terraform is not required**.

## Requirements

* [Go](https://golang.org/doc/install) or [Docker](https://docs.docker.com/install/) are required to verify the Go code compile
* (Optional) [Terraform](https://www.terraform.io/intro/getting-started/install.html) may be required to execute Unit Tests to the Terraform code

This repository/project does not store the vendors or imported Go packages, it will require all the vendors in your system if you would like to compile it using the Go in your system. If Go is not installed or you don't have the vendors in your system, the **recomended** way to verify the code compile is using the Docker container.

## Terraform Code

There is a directory for every platform and inside there is a `templates/` directory with the Terraform code that the Go code will execute.

The main files (and maybe only in the future) are:

* `provider.tf`: defines the Terraform provider access. i.e. AWS or vSphere
* `variables.tf`: defines all the Terraform variables in a minimalistic form
* `data_sources.tf`: collects data and set's variables to be used by the resources
* `resources.tf`: creates the resources that will be used with the image

Remember that the Terraform code should be simple and short because the only consumer is the Go code, not humans. Some parameters such as `description` in the variables are not required and any extra logic not related to the infrastructure should be done in the Go code, not by Terraform.

This code does not contain anything related to the Configurator, it's does not execute it nor create a file that is an input for the configurator. It's only pourpose in live is to provision a cluster in a given platform.

### Variables

If a new variable in added to the Terraform code, it has to be included in several files:

1. `templates/variables.tf`: Only the variable name, a default value if applies and type (if it's not string). Do not add description neither the string type.
2. `properties.go`: Here the variables are located in two places:
    * `Properties struct`: Include the name in snake case for every format (`json`, `yaml` and `mapstructure`) as it's in the variables.tf file. The struct field name should be in Camel case.
    * `defaultProperties`: If the variable has a default value, define it here too.
3. Also, in the KubeKit CLI, the variable has to be included in the config file.

## Unit Testing

After clonning the repository it's required to initialize the unit test environment executing:

    make unit-test-init

This is going to execute `terraform init` on every `<platform>/templates/` directory and to create the `terraform.tfvars` file with the default variables.

Edit the `<platform>/templates/terraform.tfvars` file with the appropiate values to create your cluster. **If you add new variables, these have to be added too in the Go code and the `variables.tf` file**.

To execute the unit tests for a specific platform, execute `make unit-test P=<platform>` or `make unit-test-all` to test all the platforms. The `<platform>` should be in lowercase. Example:

    make unit-test P=vsphere

This will create a cluster on the specified platform (or all). However, if you want to run some specific test using Terraform, go to `<platform>/templates/` and feel free to execute it.

To tear down the created cluster, execute `make unit-test-destroy P=<platform>` or `make unit-test-destroy-all`.

    make unit-test-destroy P=vsphere

## Build

If the changes were done to the Terraform code located in the `<platform>/templates/` directory, the Go code that encapsulate such code has to be generated. Execute `make generate`or just `make`:

    make generate

If the changes were done to the Go code there are two ways to compile:

1. If you have **Go install in your system and the vendors correctly setup**: Execute `make install` or `make test` or both `make test install`.
2. If Go is not installed but Docker is: Execute `make compile`

The safest way to proceed is the option #2 but takes more time (2-3 minutes):

    make compile

It's recommended to use an IDE such as VSCode or Atom with the Golang plugings to have a well formatted code, otherwise, execute `make fmt` before compile and push the code to GitHub.

## Use

The provisioner has the `Apply()` function to create or destroy a cluster of nodes on an infrastructure or cloud (i.e. AWS or vSphere). The `Apply()` function requires the platform name, the previous Terraform state (if any), a boolean to create or destroy the cluster and a map `map[string]interface{}` with all the parameters needed by the Terraform code to provision the cluster. The function then returns the final Terraform state of the cluster.

    import (
      "fmt"
      "github.com/kubekit/provisioner"
    )

    func Create() error {
      finalState, err := provisioner.Apply(Provisioner, Parameters, state, logger, false)
      if err != nil {
          return err
      }
      fmt.Printf("State: %s\n", finalState)
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
