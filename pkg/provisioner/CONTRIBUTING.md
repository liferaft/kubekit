# Contributing Guide

## Validating a Platform Provisioner

### Integration Test
You can run an integration test to validate all supported platforms are capable of provisioning and destroying resources. Run the following command:

`GO111MODULE=on go test github.com/kubekit/provisioner -v -run TestProvisionerIntegration -tags integration`

You can focus a specific platform by providing one of the following arguments in place of `-run TestProvisionerIntegration`:

`-run TestProvisionerIntegration/openstack` - Openstack provisioner test, depends on the following variables:

```
export OPENSTACK_AUTH_URL=http://jenga.labs.sample.com:5000
export OPENSTACK_USER_NAME=<YOUR USERNAME>
export OPENSTACK_PASSWORD=<YOUR PASSWORD>
export OPENSTACK_TENANT_NAME=<TENANT/PROJECT ID e.g. kubekit>
```

`-run TestProvisionerIntegration/aws` - AWS provisioner test, depends on the following variables:

```

```


### Build and Test in CLI
When adding a new provisioner, or modifying an existing one, you will need to know if the changes made are correct. The following manual procedure allows you to incorporate a local fork of the provisioner into KubeKit CLI (see [below](#incorporate_local_provisioner_fork)) and run a specific platform test.

#### Incorporate Local Provisioner Fork
The following procedure requires a Go v1.11 and a valid `$GOPATH` set.

1. Checkout a copy of the KubeKit CLI project into your workspace:
	- `go get -u github.com/liferaft/kubekit`
1. Change directory to the CLI project and enable the local provisioner fork:
	1. `cd $GOPATH/src/github.com/liferaft/kubekit`
	1. `make local-provisioner-on`
1. Make changes in local provisioner project
	- **Note:** kubekit CLI will expect the local copy to reside at `$GOPATH/src/github.com/liferaft/kubekit`
1. Build a KubeKit executable
	- `GO111MODULE=on go build -o ~/Desktop/kubekit-dev github.com/liferaft/kubekit`

#### Validating Openstack Provisioner in CLI
Run the following commands:

```
~/Desktop/kubekit-dev init test-cluster
```