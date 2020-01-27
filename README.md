[![Build Status](https://travis-ci.org/liferaft/kubekit.svg?branch=master)](https://travis-ci.org/liferaft/kubekit) [![codecov](https://codecov.io/gh/liferaft/kubekit/branch/master/graph/badge.svg)](https://codecov.io/gh/liferaft/kubekit) [![GoDoc](https://godoc.org/github.com/liferaft/kubekit?status.svg)](https://godoc.org/github.com/liferaft/kubekit) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# 1. KubeKit

KubeKit is a tool for setting up a Kubernetes-powered cluster.

- [KubeKit](#1-kubekit)
  - [Download](#11-download)
  - [Basic KubeKit Configuration (Optional)](#12-basic-kubekit-configuration-optional)
  - [Getting Started](#13-getting-started)
  - [Supported Platforms](#14-supported-platforms)
  - [Commands](#15-commands)
  - [The Core KubeKit Workflow](#16-the-core-kubekit-workflow)
    - [1) Create a cluster config file](#161--create-a-cluster-config-file)
    - [2.a) Edit the cluster config file](#162-a-edit-the-cluster-config-file)
    - [2.b) Set parameters with environment variables](#162-b-set-parameters-with-environment-variables)
    - [2.c) Set or export credential variables](#162-c-set-or-export-credential-variables)
    - [3) Install the kubernetes cluster](#163--install-the-kubernetes-clusterers-with-environment-variables)
    - [4.a) Provision a cluster on a cloudy platform](#164-a-provision-a-cluster-on-a-cloudy-platform)
    - [4.b) Configure Kubernetes](#164-b-configure-kubernetes)
    - [4.c) Certificates](#164-c-certificates)
    - [5) Use the Kubernetes cluster](#165--use-the-kubernetes-cluster)
    - [6) Destroy the cluster](#166--destroy-the-cluster)
  - [KubeKit Configuration](#17-kubekit-configuration)
  - [Cluster Configuration](#18-cluster-configuration)
    - [1) Platforms](#181--platforms)
      - [vSphere](#1811-vsphere)
      - [EC2](#1812-aws)
        - [How to fill the cluster configuration file for EC2](#18121-how-to-fill-the-cluster-configuration-file-for-ec2)
      - [EKS](#1813-eks)
      - [AKS](#1814-aks)
      - [Bare-metal (`raw`), Stacki and vRA](#1815-bare-metal-raw-stacki-and-vra)
    - [2.a) Node Pools and Default Node Pool](#182-a-node-pools-and-default-node-pool)
    - [2.b) TLS Keys to access the nodes](#182-b-tls-keys-to-access-the-nodes)
    - [2.c) High Availability](#182-c-high-availability)
    - [2.d) Kubernetes API access](#182-d-kubernetes-api-access)
    - [3) State](#183--state)
    - [4) Configuration](#184--configuration)
  - [Destroy the cluster](#19-destroy-the-cluster)
    - [How to manually delete a cluster](#191-how-to-manually-delete-a-cluster)
  - [Backup/Restore KubeKit Cluster Config](#110-backuprestore-kubekit-cluster-config)
    - [Backup](#1101-backup)
    - [Restore](#1102-restore)
  - [Builds](#111-builds)
    - [Development](#1111-development)
    - [Troubleshooting](#1112-troubleshooting)
  - [Integration Test](#112-integration-test)
  - [Setup Vendors](#113-setup-vendors)
    - [Go Vendor Problems](#1131-go-vendor-problems)
  - [Examples](#114-examples)
  - [KubeKit as a Service](#115-kubekit-as-a-service)
  - [Microservices](#116-microservices)

## 1.1. Download

Below are the available downloads for the latest version of KubeKit (**0.1.0**). Download the proper KubeKit binary for your operative system and architecture.

- **Mac OS X**: ( [64-bit](https://github.com/liferaft/kubekit/releases/download/v0.1.0/kubekit_0.1.0_darwin_amd64) )
- **Linux**: ( [64-bit](https://github.com/liferaft/kubekit/releases/download/v0.1.0/kubekit_0.1.0_linux_amd64) | [32-bit](https://github.com/liferaft/kubekit/releases/download/v0.1.0/kubekit_0.1.0_linux_386) )
- **Windows**: ( [64-bit](https://github.com/liferaft/kubekit/releases/download/v0.1.0/kubekit_0.1.0_windows_amd64) )

<!-- The [MD5 and SHA1 checksums](https://jfrog.com/artifactory/dependencies-snapshot-sd/uda/liferaft/kubekit/) are available online for every OS and architecture. -->

It's important for some clusters such as EKS to have KubeKit in a directory that is in the **$PATH** environment variable.

To download and install the edge version, it's required to have [Go installed](https://golang.org/doc/install) in your system. Then execute:

```bash
go install github.com/liferaft/kubekit/cmd/...
```

KubeKit and KubeKitCtl will be installed on `$GOPATH/bin/` which should be in your $PATH variable. However, this will not download the RPM with latest dependencies which may be required for most of the platforms but EKS and AKS.

The Docker images are also available in Docker Hub, to download them, it's required to have [Docker installed](https://www.docker.com/products/docker-desktop) in your system, then execute:

```bash
docker pull liferaft/kubekit:0.1.0
docker pull liferaft/kubekitctl:0.1.0
```

Or use them directly executing `docker run` like this:

```bash
docker run --rm -it liferaft/kubekit version
```

To build KubeKit and/or KubeKitCtl from source, it's required to have [Go installed](https://golang.org/doc/install). After clone the Git repo and move into the directory, use `make` to build both binaries:

```bash
git clone https://github.com/liferaft/kubekit.git
cd kubekit
make build build-ctl
```

Both binaries will be in the directory `./bin`.

## 1.2. Basic KubeKit Configuration (Optional)

You can use KubeKit with the default configuration settings, which are:

- Logs are send to the standard output and colored
- Verbose mode is enabled to print INFO, WARN and ERROR in the logs
- KubeKit home directory defaulted to `~/.kubekit.d/`

However, for better results it's recommended to configure KubeKit having your own configuration file and settings.

For more information about KubeKit configuration read the section KubeKit Configuration, for a quick configuration execute something like this:

```bash
mkdir -p ~/.kubekit.d
kubekit show-config -o yaml --to ~/.kubekit.d/config.yaml
```

In the previous instructions, feel free to select the kubekit home directory (i.e. `~/.kubekit.d`) of your preference.

Optionally, edit the file `~/.kubekit.d/config.yaml` to modify the KubeKit settings. For example, enable/disable `debug` mode or change the log file such as  `~/.kubekit.d/kubekit.log` or `/var/log/kubekit.log` or *stdout* which is the default location when `log` parameter is removed or set to `""`.

## 1.3. Getting Started

After downloading or building the binary and - optionally configuring Kubekit, the following steps are:

1. (Optional) Set or export the custom parameters and platform credential variables.
2. Create a cluster config file and - optionally - edit it to have the required parameters.
3. Install the Kubernetes cluster
4. Use the Kubernetes cluster
5. Destroy the cluster when it's not needed.

If you are in hurry, these are the commands to setup and use a cluster on **vSphere**.

```bash
# 1)
export KUBEKIT_VAR_DATACENTER='Vagrant'
export KUBEKIT_VAR_DATASTORE='sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03'
export KUBEKIT_VAR_RESOURCE_POOL='sd_vgrnt_01/Resources/vagrant01'
export KUBEKIT_VAR_VSPHERE_NET='dvpg_vm_550'
export KUBEKIT_VAR_FOLDER='Discovered virtual machine/ja186051'
export KUBEKIT_VAR_NODE_POOLS__WORKER__COUNT=3

export VSPHERE_SERVER='153.64.33.152'
export VSPHERE_USERNAME='username@vsphere.local'
export VSPHERE_PASSWORD='$up3rS3cretP@ssw0rd'

# 2)
kubekit init kubedemo --platform vsphere
# Optional: kubekit edit kubedemo

# 3)
kubekit apply kubedemo -f kubekit-2.0.0.rpm --log kubedemo.log # --debug

# 4)
eval $(kubekit get env kubedemo) # this is to export the KUBECONFIG variable
kubectl get nodes
kubectl get pods --all-namespaces

# 5)
kubekit destroy kubedemo
```

In this guide we use vSphere as example, for other platform just replace `vsphere` for the platform name and set the credentials to access platform API. Also, for a better user interface, send the logs to a file using the flag `--log` or configure KubeKit to do it as explained in [section 1.2](#basic-kubekit-configuration-optional) and optionally use `--debug` if you are having issues.

If you are using KubeKit in a shell script this is a quick example of how to use it, this time is a cluster on **AWS**:

```bash
PLATFORM=ec2
NAME=kubedemo

# 1)
echo "Remove any previous KubeKit variable"
unset $(env | grep KUBEKIT_VAR_ | cut -f1 -d=)

echo "Setting cluster parameters"
export KUBEKIT_VAR_AWS_VPC_ID='vpc-8d56b9e9'
export KUBEKIT_VAR_DEFAULT_NODE_POOL__AWS_SECURITY_GROUP_ID='sg-502d9a37'
export KUBEKIT_VAR_DEFAULT_NODE_POOL__AWS_SUBNET_ID='subnet-5bddc82c'

echo "Setting AWS credentials"
export AWS_ACCESS_KEY_ID='YOUR_AWS_ACCESS_KEY'
export AWS_SECRET_ACCESS_KEY='YOUR_AWS_SECRET_KEY'
export AWS_DEFAULT_REGION='us-west-2'

# 2)
echo "Initializing cluster configuration"
kubekit init $NAME --platform $PLATFORM

# 3)
echo "Creating the cluster"
kubekit apply $NAME -f kubekit-2.0.0.rpm

# 4)
echo "Cluster information"
kubekit describe $NAME

eval $(kubekit get env $NAME)

echo "Cluster nodes:"
kubectl get nodes

# 5)
echo "To destroy the cluster execute: kubekit destroy $NAME"
```

Modify the values of `PLATFORM` ,`NAME`, the parameters exporting the variable with prefix `KUBEKIT_VAR_` plus the parameter name and then export the platform credentials.

## 1.4. Supported Platforms

KubeKit can provision and configure Kubernetes on the following platforms:

- **VMware**, platform name: `vsphere`
- **EC2**, platform name `ec2`. This will install Kubernetes on custom EC2 instances
- **EKS**, platform name `eks`
- **Bare-metal**, platform name `raw`. It's in Beta
- **vRA**, platform name `vra`. It's in Beta
- **Stacki**, platform name `stacki`. It's in Beta and at this time behaves like `raw` platform.

## 1.5. Commands

For a complete list of commands execute

```bash
kubekit help
```

Or read the [CLI-UX](./docs/cli-ux.md) document also available [here](https://github.com/pages/liferaft/kubekit/cli.html).

## 1.6. The Core KubeKit Workflow

The core KubeKit workflow has the following 3 steps:

1. Create a cluster config file
2. Install the Kubernetes cluster

After installing the Kubernetes cluster you can:

4. Use the Kubernetes cluster
5. Destroy the cluster when it's not needed.

### 1.6.1. ) Create a cluster config file

The cluster config file contains all the required information to provision (on cloudy or hybrid platforms) a cluster and configure Kubernetes on a cluster.

Use the KubeKit subcommand `init` to generate the cluster config file with default values, the cluster name and platform are required.

```bash
kubekit init kubedemo --platform ec2
```

The cluster config file, by default, will be created in `$HOME/.kubekit.d/<UUID>/cluster.yaml`, where UUID is a 36 characters unique ID. To change the default location, export the environment variable `KUBEKIT_CLUSTERS_PATH` to the desired absolute location or with the parameter `clusters_path` in the KubeKit config file. If a relative path is set, KubeKit will get the relative path to the configuration file directory. Example:

```bash
export KUBEKIT_CLUSTERS_PATH=`pwd`
kubekit init kubedemo --platform ec2
kubekit get clusters -o wide
kubekit describe kubedemo | grep path
```

To know more about how to configure KubeKit, go to the [KubeKit Configuration](#kubekit-configuration) section.

You can get more information about the existing cluster config files with the subcommand `get clusters`. With the flag `-o wide` you'll get more information for all the existing cluster or use  `describe <cluster_name>` for a specific cluster.

### 1.6.2. a) Edit the cluster config file

Now it's time to edit the cluster config file to have the required parameters. This section is explained in detail in the [Cluster Configuration](#cluster-configuration) section, here you'll find the minimum required changes to have a working cluster on AWS.

For a quick cluster on EC2 use the following parameters as example:

```yaml
"aws_region": "us-west-2",
"aws_security_group_id": "sg-502d9a37",
"aws_subnet_id": "subnet-5bddc82c",
"aws_vpc_id": "vpc-8d56b9e9",
```

Make you have access to the `aws_vpc_id` and make sure the `aws_subnet_id` and `aws_security_group_id` are in the selected VPC.

To edit the custer configuration file, you can open the `cluster.yaml` file with the command `edit`, this will open the file with the editor defined in the variable `KUBEKIT_EDITOR`,

For example, to open the cluster config file with VS Code:

```bash
export KUBEKIT_EDITOR='/usr/local/bin/code'
kubekit edit kubedemo
```

Go to the [Cluster Configuration](#cluster-configuration) section to view some commands to help you assign or get the required values.

To view the content of the cluster.yaml file, use the flag `--read-only` or `-r` to view the file. This is also useful when you are requesting support to the KubeKit team and they request the configuration file. Make sure to remove any credential or sensitive data.

```bash
kubekit edit kubedemo -r

kubekit edit kubedemo -r | grep -v vsphere_username | grep -v vsphere_password
```

### 1.6.2. b) Set parameters with environment variables

Sometimes it's difficult or impossible to edit the file, for example, when you are using KubeKit in a bash script or creating a cluster with Jenkins. In this case, use environment variables to provide the parameters but this should be done before editing the file or use the command `update` if the configuration file already exists.

The environment variables should begin with `KUBEKIT_VAR_` followed by the parameter name. The variable is not case sensitive so could be uppercase, lowercase or mix of both.

If you look at the configuration file, in the platform or config section there are some parameters inside section or structure, for example `default_node_pool` or `node_pools` and `workers`. To assign a value to some of these parameters you have to separate them with double underscore (`__`), in some computer languages it's common to use a dot or other separator but in bash the only characters allowed are alphanumeric and underscore. For example, to set the number of CPU's in a `default_node_pool` or the number of worker nodes, use something like this:

```bash
export KUBEKIT_VAR_default_node_pool__cpus=4
export KUBEKIT_VAR_node_pools__worker__count=3
```

Other variables store a list, there are 2 ways to assign a list or array:

```bash
# Option 1: Use comma as item separator:
export KUBEKIT_VAR_time_servers="0.us.pool.ntp.org, 1.us.pool.ntp.org"

# Option 2: Use square brackets and comma:
export KUBEKIT_VAR_dns_servers="[153.64.180.100, 153.64.251.200]"
```

If an item contain space, then quote the item with single quote, like this:

```bash
export KUBEKIT_VAR_some_variable="A, B, 'C D'"
# Or
export KUBEKIT_VAR_some_variable="[A, B, 'C D']"
```

Important: Before use the environment variables, check the exported variables with `env | grep KUBEKIT_VAR_` and may be a good idea to remove them all before set them with:

```bash
unset $(env | grep KUBEKIT_VAR_ | cut -f1 -d=)
# now it's save to assign the variables
```

### 1.6.2. c) Set or export credential variables

KubeKit needs access to the platform to create/provision all the instances or VMs. In order to keep these credentials (user, password, server or keys) save out of curious eyes, the credentials could be exported variables or entered with the command `kubekit init [cluster] <name>` or with the command `kubekit login [cluster] <name>`.

 The difference between using environment variables vs the `init` or `login` command is that environment variables set global credentials, they will be used by any cluster in your system, if you have different credentials per cluster you need to update the environment variables. Using the `init` or `login` command sets the credentials for that specific cluster, there is not need to change the credentials.

Use the `login` command when the cluster configuration exists and you want to update the credentials. When the cluster is created or initialized with the `init` command it will get the credentials from:

1. The credentials flags
2. The AWS local configuration (if it is AWS or EKS)
3. The environment variables
4. Will ask to the user the missing variables (if any)

So, it's safer to provide the credentials in flags or with environment variables before using the `init` command, specially if you are using KubeKit in a script or Jenkins.

Use also the flag `--list` of the `login` command to view the credentials that KubeKit will use, like `kubekit login NAME --list`.

For **EC2** and **EKS** the variables are: **AWS_ACCESS_KEY_ID**, **AWS_SECRET_ACCESS_KEY** and **AWS_DEFAULT_REGION**

Example:

```bash
export AWS_ACCESS_KEY_ID='YOUR_AWS_ACCESS_KEY'
export AWS_SECRET_ACCESS_KEY='YOUR_AWS_SECRET_KEY'
export AWS_DEFAULT_REGION=us-west-2
```

Or using the `init` command:

```bash
kubekit login [cluster] NAME --platform PLATFORM \
  --access_key 'YOUR_AWS_ACCESS_KEY' \
  --secret_key 'YOUR_AWS_SECRET_KEY' \
  --region us-west-2
```

Or using the `login` command if the cluster was initialized:

```bash
kubekit login [cluster] NAME \
  --access_key 'YOUR_AWS_ACCESS_KEY' \
  --secret_key 'YOUR_AWS_SECRET_KEY' \
  --region us-west-2
```

For **vSphere** the variables are: **VSPHERE_SERVER**, **VSPHERE_USERNAME** and **VSPHERE_PASSWORD**

Example:

```bash
export VSPHERE_SERVER=153.0.0.101
export VSPHERE_USERNAME='username@vsphere.local'
export VSPHERE_PASSWORD='5uperSecure!Pa55w0rd'
```

Or using the `init` command:

```bash
kubekit login [cluster] NAME --platform PLATFORM \
  --server 153.0.0.101 \
  --username 'username@vsphere.local' \
  --password '5uperSecure!Pa55w0rd'
```

Or using the `login` command if the cluster was initialized:

```bash
kubekit login [cluster] NAME \
  --server 153.0.0.101 \
  --username 'username@vsphere.local' \
  --password '5uperSecure!Pa55w0rd'
```

The platforms **vRA**, **Stacki** and **Bare-metal** (`raw`) do not require to login or enter credentials because - at this time - they do not use a platform API. The user needs to enter the IP address and (optionally) the DNS name of the servers or VM's. And, either the SSH keys or the credentials to login to these servers or VM's.

Edit the cluster configuration file, locate the section `platforms.NAME.nodes` there is a list of `master` and `worker` nodes, enter the IP address on `public_ip` and the DNS (if available) on `public_dns`.

Locate the section `platforms.NAME.username` and `platforms.NAME.password` to enter the server/VM credentials. Or, if the access is password-less, enter the `platforms.NAME.private_key_file` and  `platforms.NAME.public_key_file` parameters.

Example:

```yaml
platforms:
  stacki:
    api_port: 6443
    api_address: 10.25.150.100
    username: root
    password: My$up3rP@55w0Rd
    # private_key_file: /home/username/.ssh/id_rsa
    # public_key_file: /home/username/.ssh/id_rsa.pub
    nodes:
      master:
      - public_ip: 10.25.150.100
        public_dns: master-01
      worker:
      - public_ip: 10.25.150.200
        public_dns: worker-01
      - public_ip: 10.25.150.201
        public_dns: worker-02
```

This example - at this time - is the same for **vRA**, **Stacki** and **Bare-metal** (`raw`), just replacing the platform name.

### 1.6.3. ) Install the kubernetes cluster

To have the Kubernetes cluster up and running use the subcommand `apply` like this:

```bash
kubekit apply kubedemo
```

This step, for most of the platforms, will execute two actions: provision and configure.

Some platforms or Kubernetes clusters do not allow provisioning, for example, a bare-metal cluster already exists, so it can't be provisioned with KubeKit, just configured.

### 1.6.4. a) Provision a cluster on a cloudy platform

You can skip this section if you are going to configure Kubernetes on bare-metal or an existing cluster (i.e. VRA). Go to the next section **Configure Kubernetes**.

To create an empty cluster (a cluster without Kubernetes) use the `provision` subcommand. It's required the cluster name and platform in order to locate the cluster configuration file.

```bash
kubekit apply NAME --provision
```

Example:

```bash
kubekit apply kubedemo --provision
```

The provision flag will start creating the VM's or EC2 instances, plus all other infrastructure requirements. When it's done, login to vCenter or AWS Console to view the brand-new cluster, the VM's or EC2 instances names are prefixed with the cluster name. The `provision` flag is not to configure Kubernetes, it just creates a cluster of master(s) and worker(s). With the exception of EKS and AKS, the provisioning creates a Kubernetes cluster but it's useless until you configure it.

The duration to provision a 2x2 cluster on **vSphere** is about **3 minutes** when it's cloning a template, otherwise would be around **30 minutes**. The duration to provision a 2x2 cluster on **EC2** is about **5 minutes**.

### 1.6.4. b) Configure Kubernetes

This is the process to install and/or configure Kubernetes on an existing cluster, either a cluster that was created with the `provision` subcommand or that already exists, for example, bare-metal or VRA.

The configurator requires the resource `configure` in the KubeKit cluster config file, the `init` subcommand populate it with the default parameters for the selected platform but it's recommended to double-check the parameters.

If you provisioned the cluster using KubeKit then KubeKit will set the nodes IP address and DNS at the `state` section of the cluster config file, but if you are using bare-metal or an existing cluster (i.e. VRA) then you need to provide the nodes IP address and DNS. Read the [Cluster Configuration](#cluster-configuration) section to get more information about how to do this.

Some other parameters are automatically set their values from the `provision` section or the state file such as `private_key`, `username`, `vsphere_*` and `alb_dns_name`.

Next, enter or modify the values of the other parameters such as:

- `default_ingress_host`
- `disable_master_ha`: This is located in the `platforms` section of the config file. If `true` there won't be High Availability for the Kubernetes masters. If set to `false`, you need to provide `kube_virtual_ip_api` and `kube_vip_api_ssl_port`. In platforms vSphere, Stacki and Bare-metal if you need to use an external Publi VIP to access the kubernetes API you need to provide `public_virtual_ip` and `public_virtual_ip_ssl_port` along with `public_vip_iface_name` which is the interface name on nodes where public VIP will be configured.
- `kube_virtual_ip_api` and `kube_vip_api_ssl_port`: Only if `disable_master_ha` is `false`. Make sure the IP address is available, unassigned, and reachable from wherever you are going to use Kubernetes. 
- `public_virtual_ip` and `public_virtual_ip_ssl_port`: These are available only for Stacki, vSphere and Bare-metal(raw) platforms. These are optional fields and to be used if you need to use external Public virtual IP (Public VIP) to access the kubernetes API.
When the `config` section is complete you are ready to configure Kubernetes on the cluster executing `kubekit apply NAME --configure`, for example:

```bash
kubekit apply kubedemo --configure
```

### 1.6.4. c) Certificates

Besides install and configure Kubernetes on each node, the `--configure` flag or process is going to generate TLS certificates and the `kubeconfig` file in the directory `certificates` where the cluster config file is.

If the client certificates exists they can be forced to be re-generated without having to re-apply your cluster by doing the following:

```bash
kubekit init certificates kubedemo   # regenerates certificates
kubekit apply certificates kubedemo  # applies certificates
```

If the CA certificates exists they can be forced to be re-generated but should only be done when you have reset your cluster, as we do not currently support a rolling update of the CA certificates, by using:

```bash
kubekit apply certificates --generate-ca-certs
```

The CA root certificates required to generate the key pairs (private and public certificate) will be generated as self-signed certificates unless they are provided with the following flags:

- `--etcd-ca-cert-file`: CA x509 Certificate file used to generate the etcd certificates.
- `--ingress-ca-cert-file`: CA x509 Certificate file used to generate the ingress certificates.
- `--kube-ca-cert-file`: CA x509 Certificate file used to generate the server API certificate and also, it's the generic one used by the non-provided certificates.

**EXTREMELY IMPORTANT**: It's recommended for a production cluster to provide your own signed CA certificates. In a development or testing environment it's ok to let KubeKit to generate the self-signed CA certs.

Example of how to provide your own signed CA certs:

```bash
kubekit apply kubedemo \
  --etcd-ca-cert-file string /path/to/my/ca/certs/etcd-root-ca.key \
  --ingress-ca-cert-file /path/to/my/ca/certs/ingress-root-ca.key \
  --kube-ca-cert-file /path/to/my/ca/certs/kube-root-ca.key
```

**Note**: All these `*-ca-cert-file` flags can also be used with the `--configure` flag.

If the TLS keys to access the cluster instances are not provided, KubeKit will generate them for you and store them in the `certificates` directory. All the other certificates depend of the platform where the cluster was created and will be stored in `certificates/` directory. Some certificates are specific to the instances or VMs, so they will be stored in `certificates/<hostname>/`.

You may need those certificates to login to the nodes or access Kubernetes although you don't have to login at all to the cluster instances. To use the Kubernetes API from other application (i.e. `curl` or Python script) you'll need the certificates and you will need the `kubeconfig` file to access Kubernetes with `kubectl`.

### 1.6.5. ) Use the Kubernetes cluster

When the configuration is done you need to export the `KUBECONFIG` environment variable to where the `kubeconfig` file is.

If you are watching the logs, one of the latest lines will show the location of the `kubeconfig` file. If not, you can get it with the subcommand `clusters` and flag `--describe` like this:

```bash
kubekit describe kubedemo
```

 Then, export `KUBECONFIG` like in this example:

```bash
export KUBECONFIG=~/.kubekit.d/UUID/certificates/kubeconfig
```

Then, you can verify the cluster is up and running using the `kubectl` command by doing some or all of the following commands:

```bash
kubectl cluster-info
kubectl get nodes
kubectl get pods -n kube-system
```

Now the Kubernetes cluster is ready for you. Enjoy it!

### 1.6.6. ) Destroy the cluster

When the cluster is no needed, you may want to destroy it to save money or resources. This can be done easily with the `delete cluster` subcommand, like in this example:

```bash
kubekit delete kubedemo
```

The `certificates` directory, states  `.tfstate` directory and the cluster config file `cluster.yaml` will remain, you can use them to re-create the cluster later.

To delete all the cluster generated files use the command `delete cluster-config` or `d cc` like this:

```bash
kubekit delete cc kubedemo
```

It will confirm before delete everything or you can use the flag `--force` to avoid the confirmation.

With the `delete cluster` command you use the flag `--all` that will terminate the cluster and delete all the files related to it on your system as well.

In case you want to delete all the cluster configuration files in your system, you can use the following one-liner command:

```bash
kubekit d cc $(kubekit g c -q) --force
```

Use it carefully, if one of those cluster exists, there won't be a way to recover the configuration and therefore you cannot easily destroy it.

## 1.7. KubeKit Configuration

Besides the cluster configuration file there is an optional configuration file specifically for KubeKit, but these settings could be provided to KubeKit in 3 forms:

1. A configuration file
2. Environment variables
3. Flags or parameters when KubeKit is executed

The following table shows these parameters and the parameter name for each form:

| Environment | Config File | CLI Flag    | Default     | Description |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| `KUBEKIT_CONFIG` | N/A | `--config`  | ~/.kubekit.d/config.{yaml,json}<br />./config.{yaml,json} | Location of the configuration file. |
| `KUBEKIT_DEBUG` | `debug` | `--debug` | false  | If *true*, set the highest level of logging (*debug*). Use it for development and testing, not recommended in production.  |
| `KUBEKIT_VERBOSE` | `verbose` | `--verbose` <br/> `-v` | *false*  | If set to *true*, shows more information in logs except debug information. Set log level to *info* |
|             | `log_level` |             | error | Level of detail in the logs. The possible values are: **debug**, **info**, **warning**, **error**, **fatal**, **panic** |
| `KUBEKIT_LOG_COLOR`  | `log_color` |             | true | By default, the logs are printed with colors. Set it to *false* to print them in plain text. It may be useful for processing the logs. |
| `KUBEKIT_LOG` | `log` | `--log` | *empty* == Stdout  | File to send the logs. If not set or set to an empty string, it will send the logs to Stdout, useful in Docker containers. Example: `--log /var/log/kubekit.log` |
| `KUBEKIT_CLUSTERS_PATH` | `clusters_path` |  |  | Path to store the cluster config files and assets like the certificates and state file for each cluster. |
| `KUBEKIT_TEMPLATES_PATH` | templates_path` |                        |                                                           | Path to store the template files.                            |

To generate the KubeKit config file execute the following commands:

```bash
kubekit show-config --output json --pp --to /path/to/config.json
```

You can use the output formats `json` and `yaml` for the KubeKit configuration file. Only `json` format uses the flag `--pp` for a pretty print.

For development and testing, it's recommended to set the `debug` parameter to `true`. In production or on a stable environment, you can set `debug` and/or `verbose` to `false`.

On containers, on production or when the logs won't be read by humans, you may set the `log_color` to `false`.

## 1.8. Cluster Configuration

The cluster configuration can be generated and initialized with the `init` subcommand:

```bash
kubekit init [cluster] NAME --platform PLATFORM_NAME
```

Where **NAME** is a required parameter for the name of the cluster and the object word `cluster` is optional. the flag `--platform` or `-p` is required to specify in which platform this cluster exists or will be provisioned. Example:

```bash
kubekit init kubedemo -p ec2
```

The cluster config file will be created in the directory pointed by the `clusters_path` parameter in the KubeKit config file, in `/UUID/cluster.yaml` and it will contain something like this:

```yaml
version: 1
kind: cluster
name: kubedemo
platforms:
  ec2:
    ...
    default_node_pool:
      ...
    node_pools:
      master:
        count: 1
      worker:
        count: 1
state:
  ec2:
    status: absent
config:
  etcd_initial_cluster_token: 0c3616cc-434e
  kubelet_max_pods: 110
  ...
```

There are three parameters in the root section of the document:

- `version`: It's useful to identify the version of the cluster configuration file. At this time, there is only the version 1. This version is not the KubeKit version nor tied to the KubeKit version.
- `kind`: It's used to specify what is this file for. It could be the `cluster` configuration or a `template` configuration.
- `name`: It's the name of the cluster.

Read the (1) **Platforms**, (2) **State** and (3) **Configuration** section below to know all the parameters in the resource `"platforms"`, `"state"` and `"config"` respectively.

### 1.8.1. ) Platforms

The `platforms` resource contains all the required parameters for the platform where this cluster will be created. Some parameters have default values when the `cluster.yaml` file is initialized, some required parameters are suggested and others are empty.

#### 1.8.1.1. vSphere

For example, to create a cluster in **vSphere**, use the following cluster config settings:

```yaml
datacenter: Vagrant
datastore: sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03
resource_pool: sd_vgrnt_01/Resources/vagrant01
vsphere_net: dvpg_vm_550
folder: Discovered virtual machine/ja186051
```

All these values are in the `cluster.yaml` file with the text `'# Required value. Example: <suggested value>'`.

You can optionally assign values to:

- `kube_api_ssl_port`: Port to be used by the Kubernetes API server. `kubectl` will use this port to access Kubernetes. This is the port to access a master node, not to access the VIP when HA is enabled. Default value is `6443`.
- `username`: For vSphere this should be always `root`, otherwise change it.

KubeKit will need the credentials to access the vSphere server, these credentials should be in the following environment variables: **VSPHERE_SERVER**, **VSPHERE_USERNAME** and **VSPHERE_PASSWORD**, or using the `login cluster` command.

#### 1.8.1.2. EC2

To create a cluster in **EC2**, use the cluster config settings as an example:

```yaml
aws_vpc_id: vpc-8d56b9e9
aws_security_group_id: sg-502d9a37
aws_subnet_id: subnet-5bddc82c
```

As in vSphere, all these values are in the `cluster.yaml` file with the text `'# Required value. Example: <suggested value>'`.

And add the correct values to:

- `aws_env`: This is a text appended to some resources to differentiate them from other cluster resources, make sure it's unique among the other Kubernetes clusters.
- `kube_api_ssl_port`: Port to be used by the Kubernetes API server. `kubectl` will use this port to access Kubernetes. This port is to access the API server through the ALB. Default value is `8081`.
- `username`: For AWS this should be always `ec2-user`, otherwise change it.
- `configure_from_private_net`: Set this to `true` if you are creating the cluster from an AWS EC2 instance, otherwise, i.e. from your computer, set it to `false`
- `aws_instance_placement_group`: If not empty will create all the instances in this AWS Placement Group. **IMPORTANT**: The Placement Group should exist, you need to create it, otherwise KubeKit will fail to provision the cluster.

Make you have access to the `aws_vpc_id` and make sure the `aws_subnet_id` and `aws_security_group_id` are in the selected VPC.

KubeKit will need the credentials to access EC2, these credentials should be in the following environment variables: **AWS_ACCESS_KEY_ID**, **AWS_SECRET_ACCESS_KEY** and **AWS_DEFAULT_REGION**, or using the `login cluster` command.

##### 1.8.1.2.1. How to fill the cluster configuration file for EC2

Here are some helpers to get the correct values for the EC2 parameters:

Assuming you have your AWS CLI correctly configured, to list the **VPC**'s you have access to, execute

```bash
aws ec2 describe-vpcs --query 'Vpcs[*].VpcId' --output table
```

After identifying the VPC (i.e. `vpc-8d56b9e9`), execute the following commands to list the Subnets and Security Groups in that VPC:

```bash
VPC_ID=vpc-8d56b9e9

aws ec2 describe-subnets --filter "Name=vpc-id,Values=${VPC_ID}" --query 'Subnets[*].SubnetId' --output table
aws ec2 describe-security-groups --filter "Name=vpc-id,Values=${VPC_ID}" --query 'SecurityGroups[*].GroupId' --output table
```

#### 1.8.1.3. EKS

EKS is similar in configuration to EC2. The EKS platform requires a VPC ( `aws_vpc_id`), a list of Security Groups (`cluster_security_groups`) and more than one VPC Subnets (`ingress_subnets`). Use these settings as an example:

```yaml
aws_vpc_id: vpc-8d56b9e9
cluster_security_groups:
- sg-502d9a37
ingress_subnets:
  - subnet-5bddc82c
  - subnet-478a4123
```

EKS also allows for configuration of multiple other optional variables. Use these settings as an example:

```yaml
route_53_name: ""
s3_buckets: []
kubernetes_version: "1.12"
endpoint_public_access: true
endpoint_private_access: false
cluster_logs_types:
- api
- audit
- authenticator
- controllerManager
- scheduler
```

- `route_53_name`: an optional
- `s3_buckets`: a list of s3 buckets. All nodes and pods will be granted read and write access to these buckets.
- `kubernetes_version`: the major and minor release number of a EKS supported Kubernetes release. Currently 1.12, 1.11 or 1.10. Must be a quoted value so that it is not interpreted as a decimal )
- `endpoint_public_access`: indicates whether or not the Amazon EKS private API server endpoint is enabled.
- `endpoint_private_access`: Indicates whether or not the Amazon EKS public API server endpoint is enabled.
- `cluster_logs_types`: list of logs from the eks control plane to forward to cloudwatch. Valid logs include "api","audit", "authenticator", "controllerManager" and "scheduler"

EKS Clusters currently contain three different node pools, and a default pool to set shared values.
Please note that it is currently not possible to rename node pools, or add pools beyond the three defined.

Use these settings as an example:

```yaml
default_node_pool:
  aws_ami: ami-0923e4b35a30a5f53
  kubelet_node_labels:
  - node-role.kubernetes.io/compute=""
  kubelet_node_taints:
  - ""
  root_volume_size: 100
  placementgroup_strategy: cluster
  worker_pool_subnets:
  - subnet-5bddc82c
  security_groups:
  - sg-502d9a37
```

Use these settings as an example:

```yaml
node_pools:
  compute_fast_ephemeral:
    count: 1
    aws_ami: ami-0923e4b35a30a5f53
    aws_instance_type: m5d.2xlarge
    kubelet_node_labels:
    - node-role.kubernetes.io/compute=""
    - ephemeral-volumes=fast
    root_volume_size: 100
  compute_slow_ephemeral:
    count: 1
    aws_ami: ami-0923e4b35a30a5f53
    aws_instance_type: m5.2xlarge
    kubelet_node_labels:
    - node-role.kubernetes.io/compute=""
    - ephemeral-volumes=slow
    root_volume_size: 100
  persistent_storage:
    count: 3
    aws_ami: ami-0923e4b35a30a5f53
    aws_instance_type: i3.2xlarge
    kubelet_node_labels:
    - node-role.kubernetes.io/persistent=""
    - ephemeral-volumes=slow
    - storage=persistent
    kubelet_node_taints:
    - storage=persistent:NoSchedule
    root_volume_size: 100
    placementgroup_strategy: spread
```

As in vSphere, all these values are in the `cluster.yaml` file with the text `'# Required value. Example: <suggested value>'`.

Add the correct values to:

- `username`: For EKS this should be always `ec2-user`.
- `max_pods`: This is the max number of pods the Kubernetes cluster can handle, by default it is `110`.

Make sure you have access to the provided VPC and you have access to create EKS clusters. The EKS credentials are the same as for AWS and are provided also in the same way.

#### 1.8.1.4. AKS

```yaml
# uncomment below if you want to enable preview features
# preview_features:
# - namespace: Microsoft.ContainerService
#   name: PodSecurityPolicyPreview

# the private and public key to use to access your kubernetes cluster
# set these appropriately if you want to use existing keys, otherwise they will be generated for you
private_key_file:
public_key_file:

# the azure environment
environment: public

# the resource group location you want to create resources in
resource_group_location: Central US

# if you are using an existing vnet, fill in the values below, otherwise leave blank and the vnet will be created for you
# where the name of the vnet will be taken from the cluster name
vnet_name: ""
vnet_resource_group_name: ""

# creates a private dns zone name if it is non-empty
private_dns_zone_name: ""
 
# change the vnet address space if you are using an existing vnet and set it to the same as it
vnet_address_space: 10.240.0.0/16

# define a new subnet address prefix, even if you are using an existing vnet, make sure its one that is not taken already or has overlap
subnet_address_prefix: 10.240.0.0/20

# kubernetes settings for service and docker
# do not touch if you dont know what you are doing
service_cidr: 172.21.0.0/16
docker_bridge_cidr: 172.17.0.1/16
dns_service_ip: 172.21.0.10

# set kubernetes version to empty string if you want to use the latest available in the given resource group location
# currently, you must give full major.minor.patch, ex: 1.14.8, version
# later we will support major.minor, ex: 1.14, version to be passed but that will be awhile
kubernetes_version: "" 

# the admin username to set for logging into the kubernetes worker nodes
admin_username: kubekit

# no need to change unless you know what you are doing
network_policy: calico

# pod security policy is currently in preview, if you want to try it you will need to enable to preview feature
enable_pod_security_policy: false

# the kubernetes cluster client id and secret can be set independently of the one to use create the cluster
# if its left as empty string, it will inherit the one you logged into for kubekit for provisioning
cluster_client_id:
cluster_client_secret:

# the default settings for the node pools
default_node_pool:
  # the vm instance type
  vm_size: Standard_F8s_v2
     
  # the os disk size in GiB
  root_volume_size: 100

  # the max number of pods per node, this will affect how many ip addresses azure creates for you automatically
  # do not change unless you know what you are doing and know the consequences of it
  max_pods: 30

  # the node pool type, options are: VirtualMachineScaleSets or AvailabilitySet
  type: VirtualMachineScaleSets

  # NOT IMPLEMENTED (placeholder): the docker root to change to
  docker_root: /mnt

node_pools:
  # you can define your own node pools here and override specific values if need be from the default node pool settings
  fastcompute:
    count: 3
    vm_size: Standard_F8s_v2
    root_volume_size: 100
    max_pods: 30
    type: VirtualMachineScaleSets
    docker_root: /mnt
  slowcompute:
    count: 3
    vm_size: Standard_F8s_v2
    root_volume_size: 100
    max_pods: 30
    type: VirtualMachineScaleSets
    docker_root: /mnt
```

#### 1.8.1.5. Bare-metal (`raw`), Stacki and vRA

These 3 platforms - at this time - have the same configuration and modus operandi.

For these 3 platforms there are 2 ways to get access to the nodes/servers/VM's: with SSH keys or with user/password credentials. Using SSH keys is more secure but sometimes this is not possible, so in that case, use the credentials method.

To enter the credentials or keys, use the following parameters:

- `username`: username to access the nodes. Most of the times this is `root`.
- `password`: plain text password to access the node

- `private_key_file`: absolute path to the private key file. KubeKit will create the parameter `private_key` with it's encrypted content.
- `public_key_file`: absolute path to the public key file. KubeKit will create the parameter `public_key` with it's content.

Edit, in the section `nodes`, the list of `master` and `worker` nodes. Enter the IP address on the `public_ip` parameter and the DNS (if available) on `public_dns` parameter.

And finally, select which node (or VIP if there is a Load Balancer ) will be the endpoint for the Kubernetes API server on the parameters `api_address` and `api_port` (default value is `6443`).

### 1.8.2. a) Node Pools and Default Node Pool

In every platform there is a section named `node_pools` and `default_node_pool`.

A node pool is a group if servers, nodes, instances or VMs. Each node in this node pool has the same characteristics, for example, they are all created from the same AMI, with the same memory, CPU and volume size. One of the most important parameter of a node pool is `count`, with this number you specify how many nodes with this specifications you need. Every node pool has a name, for example: `master` and `worker`, or `big_worker` , `gpu_node`, etc...

If all the node pools will have the same value for some parameters, you have to repeat the same parameter/value pair for every node pool. To avoid this, we use the <u>Default Node Pool</u>.

The default node pool contain the parameters that are applied to every node pool, unless it's specifically assigned in a node pool. For example, if all the nodes will use an specific AMI except the `big_worker`, you assign the `aws_ami` parameter inside `default_node_pool` with the AMI for every node but the big_worker nodes, and inside the `big_worker` node pool, assign the the `aws_ami` parameter with the AMI for the big_workers.

For EKS there is only one possible node group, `worker` but if you would like to use other kind of nodes such as GPU nodes, just replace the `aws_ami` and `aws_instance_type` as indicated by AWS/EKS documentation.

### 1.8.2. b) TLS Keys to access the nodes

KubeKit will use the TLS Keys you provide to access the nodes or generate them for you. This is done with the following parameters:

- `private_key_file`: A file with your own private key to access the cluster instances/VMs. If not given, KubeKit will generate it for you and store it in the `certificates` directory.
- `public_key_file`: A file with your own public key to access the cluster instances/VMs. If not given KubeKit will generated from the private key generated or given in the `private_key_file`parameter, and will be located in the `certificates` directory.

After the TLS keys are generated or used (read from the given files) KubeKit will create the following parameters. <u>**These parameters shouldn't be modified**</u>, unless you know what you are doing.

- `private_key`: Contain the private key **<u>encrypted</u>**. If a value in the config file is encrypted will be enclosed by the function-like text `DEC()` meaning, "decrypt this text before use".

- `public_key`: Contain the public key. As the public keys are meant to be distributed there is no point to have the value encrypted. This is the content of the `public_key_file` generated or provided.

Same as `DEC()` exists the function, `ENC()` meaning, "encrypt this text after use". If you like to enter the private key in the file, instead of using a private key filename, make sure to put it inside the `ENC()` function. The next time KubeKit save/update the config file, that text or private key will be encrypted and inside `DEC()` function.

### 1.8.2. c) High Availability

High Availability (HA) means that at least one master node in a cluster is available. If a master node goes down or fails, other master node will take its place. HA is not required if the cluster have only one master node, but if this node fail the entire Kubernetes cluster is not accessible.

It's possible not to have HA and have a cluster with multiple master nodes. If the master node you choose to access the Kubernetes cluster fails, you have to use other master node manually, by modifying the `kubeconfig` file. To avoid this manual change, enable HA.

By default, the Kubernetes cluster on AWS is HA, KubeKit will create an ALB that will choose an available master to serve the Kubernetes API.

On vSphere or another platform, you need to create a Virtual IP (VIP). This VIP has to be available and cannot be assigned to any instance, server or network resource.

To enable HA in your non-AWS cluster, use the following parameters:

- `disable_master_ha`: By default, it's `true` meaning the cluster is not HA (Highly Available). If you set it to `false` make sure to provide correct values to `kube_virtual_ip_api`, `kube_virtual_ip_shortname`, `kube_vip_api_ssl_port`, `public_virtual_ip`, `public_virtual_ip_ssl_port` and `public_vip_iface_name`. The VIPs should exist and be available, not assigned to any VM or network resource.
- `kube_virtual_ip_api`: Is the Virtual IP. Again, this VIP has to be available and cannot be assigned to any instance, server or network resource. For the vsphere, stacki and raw platforms this is used as an internal Virtual IP.
- `kube_vip_api_ssl_port`: Port to access the Kubernetes API server through the VIP. Cannot be the same as `kube_api_ssl_port`.
- `kube_virtual_ip_shortname`: It's a domain name assigned to the internal VIP. This is an optional value if HA is enabled.
- `public_virtual_ip`: Is the Public Virtual IP used externally to access the kubernetes API and is optional. If not required, this field should be left empty.Public virtual ip and public interface is available only for platforms vSphere, Stacki and Bare-metal(raw)
- `public_virtual_ip_ssl_port`: Port to access the Kubernetes API server through the Public VIP.
- `public_vip_iface_name`: Interface name where the Public VIP will be configured. It needs to be set as `ansible_{interface}` interface to configure the external Public VIP on interface.

### 1.8.2. d) Kubernetes API access

There are 3 parameters needed to access the API server(s). If there is no access to the API server it's not possible to access Kubernetes.

- `kube_api_ssl_port`: Port to access the Kubernetes API server. The default value and use of this port is different for each platform. Refer to each platform for more information.

- `public_apiserver_dns_name` and `private_apiserver_dns_name`: Whatever the API server is (HA with a VIP or just a single master node), you can provide a domain name for it, public and/or private.

The following are optional for public access to the API server(s) if you have configured the rest of your cluster to only be privately accessible or on a different network:

- `public_virtual_ip`: For platforms vsphere, stacki and Bare-metal(raw)) you can provide a optional public VIP that can be used to access the API externally from a node that is not in the same network. If Public VIP is not required this field should be left empty.
- `public_virtual_ip_ssl_port`: Port used to access Kubernetes API externally using Public VIP.


### 1.8.3. ) State

If you provisioned the cluster using KubeKit then KubeKit will get the nodes IP address and DNS from the state file located in the `.tfstate` directory, but if you are using bare-metal or an existing cluster (i.e. VRA) then you need to provide the nodes IP address, domain name and role name.

Open the cluster config file or execute this to open the file:

```bash
kubekit edit [clusters-config] <cluster name>
```

The state section contain the information of the cluster nodes per platform. So, inside `state` there is a platform section and it's not required to have information. So, we may find something like this:

```yaml
state:
  ec2:
    status: absent
```

Or like this:

```yaml
state:
  vsphere:
    status: running
    address: 10.25.150.186
    port: 6443
    nodes:
    ...
```

Edit the section `state.<platform>`  to enter the following parameters:

- `address`: This is the Kubernetes API server address, either a single master, Virtual IP or Load Balancer (i.e. ALB).
- `port`: Port to access the Kubernetes API server. So, the Kubernetes API server is accessible at `address`:`port`
- `status`: It's the current cluster status. You should not modify this value, KubeKit would do it, but if it's inaccurate you can manually update it. At this time, it's not very useful for KubeKit, it's just informative.
- `nodes`: Is a list of nodes in the cluster.

Each node will have the following parameters:

- `role`: It's the node role name, it should match with the Node Pool name defined in the Platform section (if there). Example of roles: `master` and `worker`.
- `public_ip`: Public IP or just the IP to access the node. You should be able to access each node with this IP, otherwise KubeKit will fail to install and configure Kubernetes.
- `public_dns`: Fully qualified domain name (FQDN) or just the hostname of the node. It doesn't have to be accessible from accessible from KubeKit. For AWS it's a FQDN accessible from KubeKit, for other platforms could be just the hostname, not accessible from KubeKit.
- `private_ip`: It's the node private IP, usable at this time only for AWS. For other platforms, it may be empty or equal to `public_ip`.
- `private_dns`: It's the private FQDN or hostname of the node. Usable at this time only for AWS, for other platforms, it may be empty or equal to `public_dns`.

The statuses of the state of the cluster are:

- `absent`: The cluster config file was created (`kubekit init`) but hasn't been provisioned yet.
- `provisioned`: The cluster exists or was provisioned with   `apply --provision`.
- `failed to provision`: The provisioning (`apply` or `apply --provision`) failed to create/provision the cluster nodes.
 `configured`: The Kubernetes cluster exists either by executing the command `apply` or `apply --configure`.
- `failed to configure`: The command `apply` or `apply --configure` failed to install or configure Kubernetes in the cluster.
- `running`: After the configured status, if the cluster is healthy, it goes to the `running` state.
- `stopped`: The cluster nodes were stopped; the cluster exist but Kubernetes is not accessible because the nodes were stopped.
- `terminated`: The cluster was deleted/terminated either using the `delete cluster` command or manually (if so, the state has to be modified manually).
- `failed to terminate`: The command `delete cluster` failed to delete the cluster.

Example:

```yaml
state:
  vsphere:
    status: running
    address: 10.25.150.186
    port: 6443
    nodes:
    - public_ip: 10.25.150.186
      private_ip: 10.25.150.186
      public_dns: kkdemoa-master-01
      private_dns: kkdemoa-master-01
      role: master
    - public_ip: 10.25.150.141
      private_ip: 10.25.150.141
      public_dns: kkdemoa-worker-01
      private_dns: kkdemoa-worker-01
      role: worker
```

### 1.8.4. ) Configuration

The configuration section have the parameters to configure Kubernetes, these parameters are platform-agnostic.

Not all the parameters required to configure Kubernetes are in this section, there are other parameters that are tie to the platform. These parameters are calculated or obtained from the platform or state section of the config file.

Some of the configuration parameters are:

- `cluster_iface_name`: The name of the network device through which Kubernetes services and pods will be communicating. If Stacki, bare metal or multi NIC generic use `ansible_byn0`. If vRA or generic (i.e. AWS, vSphere) use `ansible_eth0`.
- `public_vip_iface_name`: The network interface name where Public VIP will be configured for platforms stacki, vsphere and raw.
- `cni_ip_encapsulation`: Can be `Always` (default value) or  `Off`.
- `time_servers`: List of time servers for timesyncd
- `host_timezone`: Optional timezone configuration for host. Must be a valid zone such as "UTC", "Europe/Berlin" or "Asia/Tokyo". Will not alter host timezone settings if ommited.
- `controlplane_timezone`: Optional timezone configuration for controlplane pods ( etcd, apiserver, controller-manager and scheduler ). controlplane pods use UTC by default.
- `kubelet_max_pods`: Maximum number of pods to accept
- `docker_registry_path`: Directory where the Docker registry will store the docker images.
- `download_images_if_missing`: If `true` and an image is not in the Docker registry it will be downloaded from Docker Hub. Set this to `false` if the cluster don't have internet access.

There is also a set of parameters to configure:

- Nginx ingress: `nginx_ingress_enabled` and `nginx_ingress_controller_*`.
- Rook: `rook_enabled`, `rook_ceph_storage_*`, `rook_dashboard_*`, `rook_object_store_*` and `rook_file_store_enabled`
- etcd logs rotation: `etcd_logs_*`
- Kubernetes logs rotation: `kube_audit_log_*`

 The configuration parameters changes on every new version of KubeKit, more frequently than the platform parameters.

## 1.9. Destroy the cluster

To destroy the cluster is necessary to have the tfstate file, located in `.tfstates` directory, there is one tfstate file per platform, so they are named `<platform>.tfstate` (i.e. `ec2.tfstate`).

This means, you only can destroy a cluster that was provisioned with KubeKit.

To destroy the cluster use the subcommand `delete cluster` like this:

```bash
kubekit delete cluster kubedemo
```

### 1.9.1. How to manually delete a cluster

When the delete command fail to destroy the cluster, it has to be done manually.

To manually destroy a cluster on **vSphere**:

1. Login to the vCenter console
2. Go to the "VMs and Templates" tab
3. Go to the folder where the VM's were created. It's the `folder` parameter in the `platform.vsphere` section.
4. Select all the VM's, right click on them, and go to `Power` > `Power off`
5. Select all the VM's, right click on them, and go to `Delete from Disk`

To manually destroy a cluster on **EC2**:

1. Login to the AWS console
2. Go to `EC2` > `Instances`, select all the instances and go to `Action` > `Instance state` > `Terminate`
3. Go to `EC2` > `Key Pairs`, select the key pair named `<cluster name>_key_<aws_env>` (i.e. `kubedemo_key_aws-k8s`) and click on `Delete` button.
4. Go to  `EC2` > `Load Balancers`, select the load balancer with the name of the cluster, and go to `Action` > `Delete`.
5. Go to `IAM` > `Roles`, select or search for the roles that starts with the cluster name, select them and click on `Delete role`.
6. Open a terminal and execute:

```bash
cluster_name=

aws iam list-instance-profiles | jq -r '.InstanceProfiles[] | .InstanceProfileName' | grep $cluster_name | while read p; do echo "Deleting instance profile '$p'"; aws iam delete-instance-profile --instance-profile-name $p; done
```

The last step (#6) cannot be done in the AWS console and all the steps can be executed with the AWS CLI.

To manually destroy a cluster on **EKS**:

1. Login to the AWS console
2. Go to `EKS` > `Clusters`, select the cluster(s) to delete
3. Click on the `Delete` button at the right-upper corner.

This document do not cover how to destroy a cluster on **vRA** or provisioned with **Stacki**, for that refer to the vRA or Stacki documentation.  

## 1.10. Backup/Restore KubeKit Cluster Config

The cluster configuration for any cluster managed by KubeKit lives in a directory under `~/.kubekit.d/clusters`. This location can be changed exporting the path in the environment variable `KUBEKIT_CLUSTERS_PATH`.

### 1.10.1. Backup

To backup a cluster configuration use the command `copy cluster-config` with the flag `--export` and optionally `--zip` and `--path`. Example:

```bash
kubekit copy cluster-config kubedemo --export --zip
```

This will create a zip file in the current directory with the cluster filename. In the previous example, the file is `kubedemo.zip`.

If the `--zip` flag is not used, it will create a directory with the cluster name.

If the `--path PATH` flag is used, it will create the exported cluster as a directory or as a zip file in the specified location.

In case you have an older version of KubeKit, to backup a cluster configuration, find the directory holding the cluster with `kubekit describe NAME | grep path`, then compress that directory with `zip -r mycluster.zip <cluster directory>`.

```bash
$ kubekit describe kubedemo | grep path
  path: /Users/ca250028/.kubekit.d/clusters/9a52b458-0f11-436e-684e-331c91d7492c
$ zip -qr mycluster.zip /Users/ca250028/.kubekit.d/clusters/9a52b458-0f11-436e-684e-331c91d7492c
$
```

Or, use the following one-liner bash script:

```bash
zip -r mycluster.zip $(kubekit describe kkdemo | grep path | cut -f2 -d: | tr -d ' ')
```

### 1.10.2. Restore

A cluster configuration directory that has been backed up into a zip file can be restored and then moved into the `~/.kubekit.d/clusters/` directory.

If the zip file was create with the `copy cluster-config --export --zip` command and flags, just copy it to the `~/.kubekit.d/clusters/` directory and unzip it.

```bash
cp kubedemo.zip ~/.kubekit.d/clusters/ && cd ~/.kubekit.d/clusters/
unzip kubedemo.zip
```

If it's a directory, then copy or move the directory to `~/.kubekit.d/clusters/`

If the zip file was created with the previous one-liner, then execute following commands:

```bash
$ mkdir tmp && cd /tmp
$ unzip ../mycluster.zip
[...]
$ mv Users/ca250028/.kubekit.d/clusters/9a52b458-0f11-436e-684e-331c91d7492c ~/.kubekit.d/clusters/
```

## 1.11. Builds

Jenkins is building all the KubeKit binaries for every Pull Request. The pipeline is defined in the [Jenkinsfile](Jenkinsfile) and running on the Jenkins server (TBD). This section is to build KubeKit yourself.

The [Makefile](Makefile) is ready to build your code for your OS and several the operative systems and architectures.

Assuming you have Go installed as explained [here](https://golang.org/doc/install):

  1. Clone the repository
  2. Run `make` or `make build` to build KubeKit for your operative system and architecture

Like this:

```bash
git clone --depth=1 https://github.com/liferaft/kubekit.git && cd kubekit
make
./bin/kubekit version
```

If you don't have Go, you can build it in a Go container and all the KubeKit binaries will be in the `./pkg/{OS}/{ARCHITECTURE}/` directories with the name `kubekit`. Like this:

```bash
make build-in-docker
ls -al ../pkg/*/*/kubekit
./pkg/$OS/$ARCHITECTURE/kubekit version
```

Modify the variables `C_OS` and `C_ARCH` located at the top of `Makefile` to build binaries for different operative systems and architectures.

To remove the binaries, execute `make clean`.

To know all the actions `make` can do, execute `make help`.

### 1.11.1. Development

------

To start developing on KubeKit, execute the following steps:

1. Clone the repository:

   ```bash
   git clone git@github.com:liferaft/kubekit.git
   ```

2. Checkout the branch to modify:

   In this example, the branch to checkout is `feat/sync_configurator`.

   ```bash
   cd kubekit
   git checkout feat/sync_configurator
   ```

3. Generate the code from templates:

   If the changes were in the Terraform template (in `pkg/provisioner/*/templates/*.tf`) or the Ansible templates (in `pkg/configurator/templates/`) , it's important to generate the Go code that contain those templates. To do that, execute `make generate` on the repository.

   ```bash
   make generate
   ```

4. Build KubeKit:

   As explained in the **Build** section above, execute:

   ```bash
   make build
   ```

   Read the **Builds** section for more information.

6. Setup your playground:

   Read the Examples section for more information about how to setup your environment to use KubeKit.

   Open `kubekit/kubekit/example/config.json` and make sure the parameter `clusters_path` is set to `"./clusters"`.

   Make sure the link `kubekit/kubekit/example/kubekit` is pointing to `../bin/kubekit` and that the built kubekit binary is in `../bin`.

   ```bash
   cd kubekit/kubekit/example
   mkdir clusters
   ./kubekit version      # just to verify it's in the ../bin directory
   ```

   The last line should print: `KubeKit v1.2.4`

7. Create and destroy a cluster:

   Go to the **Getting Started** or **Examples** sections to get more information.

   First export all the AWS and vSphere credentials. There is a bug at this time that require to export them all, this will be fix and only the credentials of the platform to use will be required.

   ```bash
   export AWS_ACCESS_KEY_ID='AKIA.....................3RCQ'
   export AWS_SECRET_ACCESS_KEY='T6z......................................H4v'
   export AWS_DEFAULT_REGION=us-west-2
   export VSPHERE_SERVER=153.64.33.152
   export VSPHERE_USERNAME='username@vsphere.local'
   export VSPHERE_PASSWORD='5I9....................pc'
   ```

   Then, use these commands to create and destroy a cluster. More details in the **Getting Started** or **Examples** sections.

   ```bash
   ./kubekit init kubedemo --platform ec2
   ./kubekit get clusters
   ./kubekit edit kubedemo

   # Edit all the settings, especially those with: "Required value. Example:"

   ./kubekit apply kubedemo

   ./kubekit delete kubedemo --all
   ```

   Use other platform name instead of `ec2` to create a cluster on such platform.

### 1.11.2. Troubleshooting

To login to the nodes use the command `login node IP --cluster NAME` like this:

```bash
kubekit login node 54.202.68.123 --cluster kubedemo
```

To send or get a file to/from a node or group of nodes, use the command `copy files`.

KubeKit copy all the certificates, Ansible files and the Ansible playbook in `/tmp/kubekit/configurator`. To execute the playbook go there and execute:

```bash
cd /tmp/kubekit/configurator/
ansible-playbook -i inventory.yml -l <role_name> kubekit.yml
```

Where `role_name` is: `master000`, `master001`, `worker000`, `worker001` and so on. Try to execute the Ansible playbook on every node at same time to get the most similar results like if KubeKit is executing them.

It's also possible to execute a remote command using the KubeKit command `exec NAME --cmd COMMAND` like this:

```bash
kubekit exec kubedemo --cmd "cat /etc/kubernetes/vsphere.conf" --pools master
```

## 1.12. Integration Test

Jenkins is in charge of executing integration test every time there is a new release, however you may want to execute integration tests manually.

In the previous section (Build) is explained how to manually test to create and destroy a cluster on a platform. To do the same test in every platform, a few or just one, you can also use `make` for this.

Use the rule `test-platform-all` to create a cluster in every supported platform (EC2, vSphere and EKS) and `destroy-test-platform-all` when you are done and wants to destroy the clusters. There is also a set of rules named `test-platform-PLATFORM` and `destroy-test-platform-PLATFORM` (replace `PLATFORM` for the platform name: `ec2`, `vsphere` and `eks`) to create/destroy a cluster in such platform.

Once the cluster is created you can use `kubectl` to play with it but you can also use the rule `test-cluster` to execute a few smoke tests on the cluster to validate it's healthy and ready to use. If you want to use `kubectl` directly, remember to first execute `eval $(kubekit get env NAME)`.

Use the parameter `CLUSTER` or `C` to enter the cluster name, by default is `kkdemo`.

Set the parameter `EDIT` or `E` to `yes` or `y` to edit the configuration file before creating the cluster.

Example:

```bash
make test-platform-all
make test-cluster
eval $(kubekit get env kkdemo)
kubectl get nodes
make destroy-test-platform-all
```

Example to create a cluster on vSphere:

```bash
make test-platform-vsphere C=kubedemo E=y
make test-cluster C=kubedemo

eval $(kubekit get env kubedemo)
kubectl get pods --all-namespaces

./bin/kubekit get clusters
./bin/kubekit get nodes
./bin/kubekit login node 54.202.68.123 --cluster kubedemo

make destroy-test-platform-vsphere
```

In a different terminal, you can check the log file with:

```bash
make log-test-platform C=kubedemo
```

## 1.13. Go modules
TBD

## 1.14. Examples

Go to the `example/` directory to play or view some examples to setup a Kubernetes cluster on every supported platform.

There is a KubeKit config file to setup a verbose KubeKit and to store all the cluster config files in the `example/clusters/` directory.

There is also a link to the binary located in `bin/`, so if there isn't a binary execute `make build` to created it.

Once in the `example` directory, execute the following commands to create a Kubernetes cluster on AWS:

```bash
export AWS_ACCESS_KEY_ID='AKIA.....................3RCQ'
export AWS_SECRET_ACCESS_KEY='T6z......................................H4v'
export AWS_DEFAULT_REGION=us-west-2
```

```bash
./kubekit init kubedemo --platform ec2
./kubekit edit kubedemo
./kubekit apply kubedemo
```

```bash
./kubekit describe kubedemo
export KUBECONFIG=./clusters/<UUID>/certificates/kubeconfig
kubectl get nodes
```

```bash
kubekit delete cluster kubedemo
```

Replace `ec2` in the previous commands for `vsphere` get the same cluster on vSphere.

Go to the **Getting Started** section to get more information.

## 1.15. KubeKit as a Service

KubeKit can also be executed as a service allowing us to interact with KubeKit through a REST/HTTP API or gRPC API using mTLS (or not if disabled).

To start KubeKit as a service use the command `start` and the following options:

- `--port`: By default KubeKit runs on port `5328`, use this flag to define a different port.
- `--grpc-port`: By default the REST and gRPC API are exposed on the same port. Use this flag to make gRPC to run on a different port. The REST API will run on the port defined by `--port`.
- `--no-http`: Use this flag to not expose the REST API, only the gRPC API. It will be exposed on the default port or on the port defined by `--port`.
- `--cert-dir`: Is the directory where to locate the TLS certificates or save the generated TLS certificates. If the cert directory is not set, the default directory is `$KUBEKIT_HOME/server/pki`.
- `--tls-cert-file` and `--tls-private-key-file`: Are the location of the TLS certificate and private key. If these flags are set, the flag `--cert-dir` is ignored. If not set, the certificates would be obtained from the `--cert-dir` directory.
- `--ca-file`: Is the location of the CA certificate used to generate the server TLS certificates (if not given) or to authenticate the client certificates.
- `—insecure`: Starts KubeKit server without mTLS.

The TLS certificates are generated if they are not found or provided, these are: `kubekit-ca.{crt,key}` the CA certificate used to generate the server and client certificates, also to authenticate any client connection, `kubekit.{crt,key}` for the server and `kubekit-client.{crt,key}` for the clients.

To generate yourself the self-signed certificate use the following `openssl` commands:

- To generate the CA certificates:

  ```bash
  export TLS_PASSWD=SomeSuperPAssword
  
  openssl genrsa -des3 -passout pass:${TLS_PASSWD} -out kubekit-ca.key 4096
  openssl req -new -x509 -days 365 -key kubekit-ca.key -out kubekit-ca.crt -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=www.kubekit.io" -passin pass:${TLS_PASSWD}
  ```

- To generate the Server certificates:

- ```bash
  export SERVER_ADDR=localhost
  
  openssl genrsa -des3 -passout pass:${TLS_PASSWD} -out kubekit.key 4096
  openssl req -new -key kubekit.key -out kubekit.csr -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=${SERVER_ADDR}" -passin pass:${TLS_PASSWD}
  
  openssl x509 -req -days 365 -in kubekit.csr -CA kubekit-ca.crt -CAkey kubekit-ca.key -set_serial 01 -out kubekit.crt -passin pass:${TLS_PASSWD}
  
  openssl rsa -in kubekit.key -out kubekit.key.insecure -passin pass:${TLS_PASSWD}
  mv kubekit.key kubekit.key.secure
  mv kubekit.key.insecure kubekit.key
  ```

- To generate the Client certificates:

  ```bash
  openssl genrsa -des3 -passout pass:${TLS_PASSWD} -out kubekit-client.key 4096
  openssl req -new -key kubekit-client.key -out kubekitctl.csr -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=${SERVER_ADDR}" -passin pass:${TLS_PASSWD}
  
  openssl x509 -req -days 365 -in kubekitctl.csr -CA kubekit-ca.crt -CAkey kubekit-ca.key -set_serial 01 -out kubekit-client.crt -passin pass:${TLS_PASSWD}
  
  openssl rsa -in kubekit-client.key -out kubekit-client.key.insecure -passin pass:${TLS_PASSWD}
  mv kubekit-client.key kubekit-client.key.secure
  mv kubekit-client.key.insecure kubekit-client.key
  ```

To start the server use the command `start server`, not using any option at all will generate the certificates on `$KUBEKIT_HOME/server/pki`, gRPC and REST API exposed on the same port `5328` as well as the Healthz service.

```bash
kubekit start server
```

As REST API is running, you can use `curl` to access the KubeKit or the Healthz service:

```bash
$ curl -s -k -X GET https://localhost:5823/api/v1/version | jq
{
  "api": "v1",
  "kubekit": "2.1.0",
  "kubernetes": "1.12.5",
  "docker": "18.06.2-ce",
  "etcd": "v3.3.12"
}
$ curl -s -k -X GET https://localhost:5823/healthz/v1/Kubekit | jq
{
  "code": 1,
  "status": "SERVING",
  "message": "service \"v1.Kubekit\" is serving",
  "service": "v1.Kubekit"
}
```

To access the gRPC API, temporally, use the `kubekitctl` command:

```bash
$ kubekitctl -cert-dir $HOME/.kubekit.d/server/pki version
Health Check Status for service "v1.Kubekit":
  GRPC: SERVING
  HTTP: SERVING

Version:
  gRPC Response: {"api":"v1","kubekit":"2.1.0","kubernetes":"1.12.5","docker":"18.06.2-ce","etcd":"v3.3.12"}
  HTTP Response: {"api":"v1","kubekit":"2.1.0","kubernetes":"1.12.5","docker":"18.06.2-ce","etcd":"v3.3.12"}
```

The `kubekitctl` command is a work in process as well as the KubeKit server.

## 1.16. Microservices

Go to the [KubeKit Microservices Example](https://github.com/liferaft/kubekit-micro-examples) to use KubeKit as a microservices application.
