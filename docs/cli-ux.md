# CLI UX design and proposal

This is a proposal for a new design of the Command Line Interface (CLI) commands in order to improve the user eXperience (UX), targeted primarily for human readability and usability.

A command language is the part of the CLI with which the user interacts with a program, in this case KubeKit. At this time the only way to interact with KubeKit is the CLI.

## Table of contents

<!-- TOC -->

- [CLI UX design and proposal](#cli-ux-design-and-proposal)
  - [Table of contents](#table-of-contents)
  - [Command Structure](#command-structure)
    - [Verbs](#verbs)
    - [Nouns](#nouns)
    - [Basic Commands](#basic-commands)
      - [`help`](#help)
      - [`version`](#version)
      - [`show-config`](#show-config)
      - [`completion`](#completion)
      - [`token`](#token)
    - [Parameters](#parameters)
  - [KubeKit Commands](#kubekit-commands)
    - [`init`](#init)
      - [Initialize a `cluster`](#initialize-a-cluster)
      - [Initialize a `template`](#initialize-a-template)
      - [Create `certificates`](#create-certificates)
      - [Create a `package`](#create-a-package)
    - [`apply`](#apply)
      - [Apply changes to a `cluster`](#apply-changes-to-a-cluster)
      - [Apply a `package` to a cluster](#apply-a-package-to-a-cluster)
    - [`delete`](#delete)
      - [Delete a `cluster`](#delete-a-cluster)
      - [Delete a `cluster configuration`](#delete-a-cluster-configuration)
      - [Delete a `template`](#delete-a-template)
      - [Delete a `file`](#delete-a-file)
      - [Delete a `package`](#delete-a-package)
    - [`edit`](#edit)
    - [`get`](#get)
      - [Get `clusters`](#get-clusters)
      - [Get `nodes`](#get-nodes)
      - [Get `templates`](#get-templates)
      - [Get `environment`](#get-environment)
    - [`copy`](#copy)
      - [Copy a `cluster`](#copy-a-cluster)
      - [Copy a `cluster configuration`](#copy-a-cluster-configuration)
      - [Copy a `template`](#copy-a-template)
      - [Copy a `file`](#copy-a-file)
      - [Copy `certificates`](#copy-certificates)
      - [Copy a `package`](#copy-a-package)
    - [`exec`](#exec)
      - [Execute a command or script file on a `cluster`](#execute-a-command-or-script-file-on-a-cluster)
      - [Execute or install a previously copied `package`](#execute-or-install-a-previously-copied-package)
    - [`login`](#login)
      - [Login or enter credentials of a `cluster`](#login-or-enter-credentials-of-a-cluster)
      - [Login to a `node`](#login-to-a-node)
    - [`describe`](#describe)
      - [Describe `clusters`](#describe-clusters)
      - [Describe `templates`](#describe-templates)
      - [Describe `nodes`](#describe-nodes)
      - [Describe `packages`](#describe-packages)
    - [`start`, `stop` and `restart`](#start-stop-and-restart)
      - [Start/Stop `cluster`](#startstop-cluster)
      - [Start/Stop `server`](#startstop-server)
    - [`scale`](#scale)
  - [Implementation matrix](#implementation-matrix)

<!-- /TOC -->

## Command Structure

All the KubeKit commands should follow the standard: *VERB NOUN PARAMETERS*

All the commands (verb, noun or adjective) should always be a single, lowercase word without space, underscores or other word delimiter. However if there is not obvious way to avoid having multiple words, separate with kebab-case, like `cluster-config`.

### Verbs

The verbs or actions KubeKit cover are:

- `init` or `i`
- `apply` or `a`
- `delete` or `d` or `del` or `destroy`
- `edit` or `e`
- `get` or `g`
- `copy` or `cp`
- `exec` or `x`
- `login` or `l`
- `describe` or `desc`
- `start`
- `stop`
- `restart`
- `scale`

Some verb have a short-named version.

### Nouns

Nouns (or objects) can be in both the singular or plural form. Unless the singular form is the clearer choice, the language design should err on the side of using the plural form. For example, most of the verbs for `cluster` acts to only one single cluster, so the noun is in singular form.

These nouns are as follow:

- **`cluster`** * or `c`
- `clusters-config` or `cc`
- `templates` or `t`
- `nodes` or `n`
- `files` or `f`
- `certificates` or `certs`
- `package` or `pkg`
- `environment` or `env` or `e`
- `server` or `srv` or `s`

The default noun is `cluster` (marked with *) but an action is required. Some of the nouns have a short-named version.

### Basic Commands

There are some keywords that are just commands. They could be identified as verbs or nouns but they exists in most of the CLI for many commands or products. These are:

- `help` or `h`
- `version`
- `show-config`
- `completion`
- `token`

#### `help`

Help a description of what the given command does. It also prints a list of sub-commands and a description for each one. It will show too the available flags and a short description about them.

```bash
kubekit help [command-name]
```

#### `version`

Version prints the name of the program ("KubeKit") and the current version.

```bash
kubekit version \
  --verbose
```

With the `--verbose` or `-v` flag prints the version, build number, git hash and version of Kubernetes, Docker and etcd used for this KubeKit release. Also, the latest version if there is internet access to the Github or Artifactory.

#### `show-config`

A user may use this command to auto-generate the KubeKit configuration file. Also, it could be used to know what are the configuration settings used by KubeKit.

```bash
kubekit show-config \
  --format json|yaml|toml \
  --pp
  --to FILE
```

The `--format` or `-f` flag is to print the configuration settings in different format, by default it is JSON. If JSON is the selected format, use `--pp` or `-p` for pretty print it, using indent.

Use the flag `--to` to save the configuration into a file. This command with this flag may be the first command executed after download KubeKit to generate a default configuration file to be customized later.

#### `completion`

Completion generate the Bash completions for KubeKit. The output of this command should be executed or included into the `~/.bashrc` or `~/.profile` files.

```bash
. <(kubekit completion)
```

```bash
kubekit completion >> ~/.bashrc
```

#### `token`

Token is a temporal command created just for EKS to mimic the command `aws-iam-authenticator token` used to grant access to the EKS cluster. When the need of this command is not required by EKS this command will be removed from KubeKit.

This command was implemented to eliminate the need to install `aws-iam-authenticator` and create an external dependency of KubeKit.

```bash
kubekit token -i CLUSTER-NAME
```

### Parameters

With every command (verb + noun) there may be a set of parameters, arguments or adjectives. These parameters usually are flags that follows the [GNU coding standards](https://www.gnu.org/prep/standards/html_node/Command_002dLine-Interfaces.html).

The flag parameters have 2 forms:

- **long-named parameter** are prefixed by two dashed and are formed by single or multiple words using kebab-case format.
- **short-named parameters** are prefixed by a single dash and usually is just one lower-case letter. However, there may be parameters with one upper-case letter or two letters.

We used the flags to name the parameters and make the command clearer, so flags are preferred to arguments.

The arguments are the basic way to provide input to a command and while flags are preferred, they are sometimes unnecessary when there is only one possible argument or it's obvious. For example, `kubekit init cluster`, the main and next parameter here is the cluster name, instead of use `--cluster-name` to provide this parameter, let's just enter the cluster name after it, like `kubekit init cluster some-cluster-name`.

There are some standard parameters, some are global to every parameter (persistent) and other are specific to a command. For example:

- `--help` or `-h` is a parameter of every command and it's used to print how to use such command. There is also the command `help`. So, it could be used in 3 forms.
- `--version` is a parameter of the root command and also a command (it's a noun with no verb or the default verb) `version` to print the KubeKit version. If used with the `--verbose` or `-v` flag will provide detailed information about the version.
- `--scroll` is a persistent command to make the output to scroll up or print the tasks in the same location
- `--verbose` or `-v` is a persistent command to provide more information about the executed command.
- `--quiet` or `-q` is a persistent command to print nothing to the screen, except if there is a critical error.
- `--no-color` is a persistent command to print all the output with the terminal default colors. By default KubeKit will print all the output with colors.

## KubeKit Commands

### `init`

Init is used to initialize, generate or create a cluster configuration file, template, certificates or package. This is usually the first command to execute. When it's initializing a cluster the noun cluster is optional, as it's the default noun.

#### Initialize a `cluster`

```bash
kubekit init [cluster] CLUSTER-NAME \
  --platform cluster-platform \
  --path config-file-location \
  --format json|yaml|toml \
  --template template-name \
  --update \
  --access_key aws_access_key_id \
  --secret_key aws_secret_access_key \
  --region aws_default_region \
  --profile aws_profile \
  --server server_ip_or_dns \
  --username username \
  --password password
```

The cluster requires a platform where it's going to exists (e.g. `ec2` or `vsphere`), and it's specified with the flag `--platform` or `-p`.

The init cluster command creates the cluster configuration file in the default location (`~/.kubekit.d`) or in the directory specified with `--path` flag.

By default the cluster configuration file is a YAML file but could be JSON or TOML if the format is specified with the flag `--format` or `-f`.

A cluster can be created from a template with the `--template` or `-t` flag. This flag specifies the template name or location.

For **EC2** and **EKS** please see the [`login`](#login) command for the description of the Amazon options.

An existing configuration cluster file - by default - cannot be overwritten, once it's created it with `init` it cannot be re-created with `init`. To update the file you have to use the `edit` command to modify the parameters. However, if you use the `--update` flag KubeKit will overwrite the cluster configuration with the values set in the environment variables starting with `KUBEKIT_VAR_` plus the parameter name. This flag is useful when using KubeKit with a script to automate a process.

This command also generates the credentials for the platform. This could be done with the flags or with environment variables. Read the `login` command for more information about these flags or environment variables for credentials.

#### Initialize a `template`

```bash
kubekit init template template-name \
  --platform platform[,platform...] \
  --path config-file-location \
  --update
```

A template is a file with a specific cluster configuration for  platforms. By default it will create the default configuration for all the platforms but you can specify which one with the flag `--platform` or `-p`. The templates are used to create/initialize a cluster based on the template configuration for one of the template platforms.

The templates are also stored in a default location  (`~/.kubekit.d/templates`)  but can be also stored somewhere else with the flag `--path`.

This command also have the `--update` flag to overwrite the template taken the parameters from environment variables starting with `KUBEKIT_VAR_`. By default the `init template` command will fail if you are trying to initialize an existing template, unless you use the `--update` flag.

#### Create `certificates`

It's required to initialize the cluster prior the creation of certificates.

```bash
kubekit init certificates CLUSTER-NAME \
  --etcd-ca-cert-file /path/to/my/ca/certs/etcd-root-ca.key \
  --ingress-ca-cert-file /path/to/my/ca/certs/ingress-root-ca.key \
  --kube-ca-cert-file /path/to/my/ca/certs/kube-root-ca.key
```

The cluster certificates can be created in advance of the cluster creation or before the configuration, if not they will be created as part of the cluster creation process with the `apply` command.

If the CA certificates are not provided, KubeKit will create them for you. Notice that it's highly recommended  to create or provide your own CA certificates for production environments or environments with some sort of access restriction.

The CA certificates can be provided to KubeKit with a flag for each CA cert, these flags start with the cert name, then `-ca-cert-file`. Like:

- `--kube-ca-cert-file`: CA certificate for Kubernetes API server. It's also the generic CA certificate .If only this certificate is provided, then all the other public certificates will be generated from this one.
- `--etcd-ca-cert-file`: CA certificate for etcd
- `--ingress-ca-cert-file`: CA certificate for Kubernetes Ingress

#### Create a `package`

```bash
kubekit init package [CLUSTER-NAME] \
  --format (rpm|deb) \
  --target filename
```

Creates a package RPM or DEB for the provided cluster name. The format, by default, is RPM unless the flag `--format` or `-f`  is used.

The package file will be located in the cluster directory unless the flag `--target` or `-t` is used to indicate the filename or directory to store the package. The default package name is `kubekit.rpm` or `kubekit.deb`

*TBD: Still is pending how to provide the manifest or list of software and version to include in the package.*

### `apply`

Apply is used to apply the configuration into your infrastructure. This means `apply` will create the cluster if it doesn't exists or will update or apply the configuration changes if it exists.

The nouns linked to `apply` are a cluster and a package (i.e. an RPM or DEB file).

#### Apply changes to a `cluster`

```bash
kubekit apply [cluster] cluster-name \
  --provision \
  --configure \
  --package-file FILE \
  --force-pkg \
  --generate-certs \
  --etcd-ca-cert-file /path/to/my/ca/certs/etcd-root-ca.key \
  --ingress-ca-cert-file /path/to/my/ca/certs/ingress-root-ca.key \
  --kube-ca-cert-file /path/to/my/ca/certs/kube-root-ca.key \
  --export-tf \
  --export-k8s \
  --plan
```

KubeKit does three main things to have a Kubernetes cluster running: (1) provision, (2) generate certificates and (3) install and configure Kubernetes and related services.

Some of these 3 actions may be optional in some scenarios for example if the cluster (group of nodes, VMs or instances) already exists the provisioning is not done, when the certificates where created previously by the user or KubeKit, it will not create them again, and for EKS or AKS clusters there is none of these 3 actions, just create/update the Kubernetes cluster.

So, in some cases the user may want to execute each action separately. For that we have the flags `--provision`,  `--configure` and `--certificates`. The flag `--provision` or `-p` is going to provision (if possible for the cluster platform) the cluster, it will create or update the nodes, VMs or instances of the cluster. 

If in the cluster directory there is a file named `kubekit.rpm` or  `kubekit.deb` the package will be uploaded to every cluster node and installed before doing the configuration. This package file contain Docker images required by Kubernetes. Use the flag `--package-file` if the default package file is not in the cluster directory. Some platforms or scenarios requires to use the `--force-pkg`, this flag is to force the installation of the package even on platforms that a package does not make sense such as EKS and AKS.

With the flag `--configure` or `-c`, KubeKit will generate the certificates (if doesn't exists), install and configure Kubernetes on the provisioned or existing cluster.

The flag `--certificates` will renew the cluster certificates with the existing certificates, unless `--generate-certificates` is used. If the certificates does not exists, they will be created. To apply certificates the Kubernetes cluster must exists. *It's under discussion to keep or remove this feature as it can be done with `kubectl`*

The `--generate-certs` flag will force KubeKit to generate the certificates even if they already exists. And the `--*-ca-cert-file` flags will provide the path to the CA certificates the user wants to use to generate the public certificates. Just like in the command `init certificates`.

The following commands are for advance users so they are not displayed in the help neither have a short name.

The `--export-tf` flag will do nothing but to create the Terraform templates or code so you can provision the infrastructure with Terraform. This is kind of handy when for some reason KubeKit is failing to provisioning and Terraform doesn't. There is also a `--export-k8s` to export the Kubernetes manifests that will be applied to the cluster once it's up and running. This is useful to delete, modify or re-apply the created resources. 

The `--plan` flag is to print the changes that will be applied to the infrastructure, but nothing will be really done. 

#### Apply a `package` to a cluster

```bash
kubekit apply package CLUSTER-NAME \
  --package-file FILE \
  --backup
```

Sometimes, mostly on development or testing, it's needed to use a different version of Kubernetes, Docker or another software or Docker image, this is when `apply package` commands is useful, it uploads a package (i.e. an RPM or DEB file) to every node of the cluster and install it.

By default the package will be taken from the cluster directory and it's named `kubekit.rpm` or `kubekit.deb`, unless the flag `--package-file` is used to provide the package file.

If there is a previously copied package in the node, it will be overwritten unless the flag `--backup` is used. This command do not remove the package file from the node once it's installed. The target directory at the node is `/tmp/`.

### `delete`

Delete is similar to `apply` when the noun is a cluster but in the opposite way. It will terminate all the cluster nodes and may also delete the cluster configuration. Destroy also work with most of the nouns: cluster-config, templates, file and packages. It's on discussion to include: node and certificates.

All the delete commands will confirm with the user before destroy the object. In some cases this is not wanted and that's why exists the `--force` flag, to force KubKit to delete the object and do not ask questions.

#### Delete a `cluster`

```bash
kubekit delete [cluster] cluster-name \
  --all \
  --force
```

This command will destroy all the cluster nodes if the platform allows it. For example, this is possible on EC2 and vSphere but not for bare-metal or vRA.

If the flag `--all` is set, it will also delete the all cluster configuration files such as the config file, certificates and latest state.

#### Delete a `cluster configuration`

```bash
kubekit delete cluster-config CLUSTER-NAME[,CLUSTER-NAME ...] \
  --force
```

Will delete all the cluster configuration files only if the cluster does not exists, unless `--force` is used.

It's important to know that if the cluster exists and KubeKit is forced to delete the configuration there is no way for KubeKit later to modify or access this cluster. So, be careful with this command.

#### Delete a `template`

```bash
kubekit delete template template-name[,template-name ...] \
  --force
```

It just deletes the given template.

If the template is located somewhere else (not in the default location) you just can delete the template file.

#### Delete a `file`

```bash
kubekit delete files CLUSTER-NAME file[,file ...] \
  --nodes node[,node ...] \
  --pools pool-name[,pool-name ...] \
  --force
```

This will delete a file on every node of the cluster or on specific nodes listed with `--nodes` , or `-n`, or on the specific pools listed on `--pools` or `-p`.

The name of the nodes could be the IP address, DNS or the node ID which is the pool name and 3 digits. It's allowed to use wildcards like `?` and `*` and ranges like `[1-3]`.

It may also delete the file from all the nodes that belongs to a node pool which is specified with the flag `--pool`, the name of the pools also allow wildcards and ranges, like the nodes.

#### Delete a `package`

```bash
kubekit delete packages CLUSTER-NAME \
  --force
```

Deletes the package from the cluster directory. The package filename is `kubekit.rpm` or `kubekit.deb`.

If the file was created in a different location, just delete the file.

### `edit`

Edit is basically for a cluster configuration but as it's an action for only one object it can be ignored and threaded as the default object.

```bash
kubekit edit [clusters-config] cluster-name[,cluster-name ...] \
  --editor /path/to/editor \
  --read-only
```

The edit command is going to open an editor with the configuration file for each given cluster name. The editor used will be taken from the flag `--editor` or `-e`, if not provided will be taken from the environment variable `KUBEKIT_EDITOR` and if this variable is not exported the used editor is `/usr/bin/vi`. If `vi` is not there, then KubeKit will throw an error.

The `--read-only` or `-r` flag is to show to the standard output the content of the file. It's like `cat` the file. If more than one cluster is given, they will be printed one after the other.

This is a useful command and usually the command to execute after initializing the cluster because the initialization will create a cluster configuration file with default values but other parameters do not have values assigned.

### `get`

The `get` command is used to retrieve information from a given object. If no object name is specified, it will return information from all the occurrences of such object. The `get` command mimics the behavior of the GET method of a REST API.

For example, `get nodes` list all the known nodes in a given cluster with information about those nodes. However, if the flag `--node node-A` is used, KubeKit will retrieve information only about the node `node-A`.

It is used with most of the objects: clusters, nodes, files and templates. *Note: certificates are currently being discussed for inclusion*.

The flag `--output` or `-o` is a persistent flag, this means it can be used with any object. This flag is used to request the output in a specific format or with more information. The possible values are:

- none: none or not include `--output` flag will print the regular information about the object in a table or human readable format.
- `wide` or `w`: it's like none but provide more information.
- `json`: prints the output in JSON format in a single line. It can be used with the flag `--pp` to pretty print the JSON output, using two space indentation.
- `yaml`: prints the output in YAML format.
- `toml`: prints the output in TOML format.

JSON, YAML and TOML provide the same amount of information as like `wide` but in the required format.

However, if the global parameter `--quiet` or `-q` is in use then the output will be just the names or ID of the requested objects. So, using quite mode and an output format at same time will throw an error.

#### Get `clusters`

This action actually applies to clusters configuration file in the default location but in order to make it simpler, the noun is replaced to `clusters` instead of `clusters-config`.

It's used to print information about the clusters configured in my system, like: cluster name, total number of nodes, status and platform. If the parameter `--output` or `-o` is set to `wide` or `w` will also print the location of the configuration file and the Kubeconfig file if exists. Use the `--quiet` or `-q` flags to print only the name of the clusters.

The option `--kubeconfig` is to just print the location of the Kubeconfig file for the given clusters. If the cluster does not exists will print `N/A`.

```bash
kubekit get clusters NAME[,NAME...]\
  --kubeconfig \
  --output wide|json|yaml|toml \
  --pp
```

#### Get `nodes`

This command is very similar to `kubectl get nodes` but `kubectl` prints the nodes of an existing cluster and `kubekit` print the cluster nodes even if the cluster doesn't exist yet.

So, it prints the list of node from the given cluster name with the following information: name, public IP address, pool name and status. If the `wide` or `w` option is set on `--output` it will also print: age, private IP, private and public DNS.

```bash
kubekit get nodes CLUSTER-NAME[,CLUSTER-NAME...]\
  --node node[,node ...] \
  --pool pool-name[,pool-name ...] \
  --output wide|json|yaml|toml \
  --pp
```

If the flags `--node` or `-n` are present, the entire nodes list will be printed, or a partial list matching the given expression, along with information about these nodes. This expression can use wildcards (`?`, `*`) and ranges, (i.e `[1-5]`).

Same with the flag `--pool` or `-p` to print the nodes that belong to the listed pool names or match with the pool name expressions.

#### Get `templates`

Get templates is similar to `get clusters` but instead of list the cluster configuration files will list all the templates located in the default location.

```bash
kubekit get templates NAME[,NAME...]\
  --output wide|json|yaml|toml \
  --pp
```

The printed information with no output (none) is: template name, total number of nodes and supported platforms. Using the `wide` or `w` output will also print: node pools.

#### Get `environment`

Get environment, or env, prints out export commands which can be run in a subshell. Use this command with the shell command `eval` to export or set the environment variables required to work with the given cluster.

This command works for different shells. By default it prints the export commands for bash, but if you are in a different shell use the flag `--shell`. The supported shells are: `bash`, `fish`, `cmd` (Windows) and `powershell`.

```bash
kubekit get env NAME
  --shell SHELL
```

Use the flag `--unset` or `-u` to prints out unset commands which reverse the command effect (in such case, the cluster name is not required).

```bash
kubekit get env --unset
```

### `copy`

The copy command applies to the nouns: clusters, clusters configuration, templates, files, certificates and packages.

#### Copy a `cluster`

Copy a cluster is going to duplicate an existing cluster in the same platform with a different name. It's like executing the `apply cluster` command after coping the cluster configuration with a new cluster name.

```bash
kubekit copy [cluster] NAME \
  --to new-cluster-name \
  --provision \
  --configure \
  --generate-certs \
  --etcd-ca-cert-file /path/to/my/ca/certs/etcd-root-ca.key \
  --ingress-ca-cert-file /path/to/my/ca/certs/ingress-root-ca.key \
  --kube-ca-cert-file /path/to/my/ca/certs/kube-root-ca.key \
  --export-tf \
  --export-k8s \
  --plan
```

All the flags have the same description as in the `apply cluster` command except for `--to` which is the name of the new cluster after coping the configuration file.

If you want to copy or duplicate an existing cluster in a different platform, first copy the cluster configuration providing this platform, edit the platform parameters and apply the changes. So, it's not possible to duplicate a Kubernetes cluster into a different platform without editing the config file first.

#### Copy a `cluster configuration`

Copying a cluster configuration is to export or duplicate a configuration file with a new cluster name. The result can go to the default KubeKit clusters directory, a given directory in `--path`, the current directory (if `--export` is used without `--path`) or to a Zip file.

```bash
kubekit copy cluster-config CLUSTER-NAME \
  --export \
  --zip \
  --to new-cluster-name \
  --platform platform \
  --path /path/to/new/template/file \
  --format json|yaml|toml \
  --template template-name
```

If the flags `--platform` or `--template` are used, it will create a copy of the cluster configuration and then execute the `init cluster` command to provide default parameters for the new platform or use the parameters from the given template.

If `--path` is not used, the cluster configuration file will be stored in the default location.

If `--export` is used the cluster will be copied with the same name to the given path. If the path is not provided the cluster will be exported to the current directory. Use the `--zip` flag with export to create the copy in a zip file, using the cluster name as filename.

The flag `--export` cannot be used with `--platform` or `--template`.

The `--zip` flag is to be used only with `--export`.

#### Copy a `template`

This command creates a copy of an existing template with a different name.

```bash
kubekit copy template NAME \
  --to new-template-name \
  --platforms platform[,platform ...] \
  --path /path/to/new/template/file
```

If the `--platform` flag is used, the new template will be created for the given platforms. For the listed platforms that exists in the source template, the same configuration will be used. For the listed platforms that are not in the source template, the default values will be used.

If `--path` is not used, the template will be stored in the default location.

#### Copy a `file`

Copy a file is like `scp` to transfer files from the local host to an specific location at the selected nodes.

```bash
kubekit copy files CLUSTER-NAME \
  --from [host:]/path/to/files \
  --to [host:]/path/to/files \
  --nodes node[,node ...] \
  --pool pool-name[,pool-name ...] \
  --force \
  --backup \
  --sudo \
  --owner owner \
  --group group \
  --mode mode
```

The copy command use the following flags:

- `--from` or `-f`: specifies the location of the source file or files to copy/transfer to the target nodes. Usually this file is on localhost but it's possible to get a file from a remote host, just make sure KubeKit have access to this host with the provided or generated SSL keys. By definition the source file(s) should be located in  only one location, so KubeKit cannot get files from different hosts.
- `--to` or `-t`: specifies the path where to put the transferred file. If the target host is just one it can be specified here, otherwise use the `--nodes` or `--pool` flags. If the goal is to get a file from remote hosts to localhost, do not specify host, nodes not pools.
- `--nodes` or `-n`: specifies the list of hosts or nodes where to put the file in the path given with `--to`. A node name of the list could be an IP, DNS or node name, and these could use wildcards (`?`, `*`) or a range (`[1-3]`).
- `--pool` or `-p`: specifies a list of pool names. The file(s) will be transferred to all the nodes that belong to each pool.
- `--force`: By default if the file exists at the target host and path, the file won't be transferred. Use the force flag to make KubeKit to overwrite the file if it already exists.
- `--backup` or `-b`: creates a backup file if it already exists.
- `--sudo`: let KubeKit to use the `sudo` command to create the backup or change owner or group. The user used to access the nodes need to have sudo access.
- `--owner` or `-o`: specifies the name or uid (user id) of the user that will own the target file. The user to access the nodes must have sudo access.
- `--group` or `-g`: specifies the name or gid (group id) of the group that will own the target file. The user to access the nodes must have sudo access.
- `--mode` or `-m`: specifies the mode of the target file. The mode has to be an octal number (`0644` or `01755`) prefixed by zero.

Make sure the username provided to KubeKit has permissions to copy the file to the desired location/path or to read it from the remote location.

The flags `--owner`, `--group` and `--sudo` are not used when coping files from a remote server to localhost.

#### Copy `certificates`

This command is used to import or export certificates to/from a cluster. After the certificates are imported they will be applied into the cluster, just like the `apply` command with the flag `--generate-certificates`.

```bash
kubekit copy certificates \
  --from [host:]/path/to/cert/files \
  --from-cluster cluster-name \
  --to [host:]/path/to/cert/files \
  --to-cluster cluster-name \
  --dry-run \
  --zip \
  --no-backup
```

If the flag `--dry-run` is used the to import certificates into a cluster, the certificates won't be applied in the cluster.

The exported certificates will be stored in the given path in a directory named like the cluster name. If the flag `--zip` is used they will be in a zip file named like the cluster name.

If there are some certificates at the target path, they will be backed up unless the flag `--no-backup` is used.

#### Copy a `package`

The copy package command is used to transfer a package file to every cluster node. It's similar to copy a file but with less parameters.

```bash
kubekit copy package CLUSTER-NAME \
  --package-file FILE \
  --backup
```

If the package is not in the default location use the flag `--package-file` or `-f`. If there is a previously copied package in the node, it will be overwritten unless the flag `--backup` is used.

This command will not install the package, to do so use the command `exec package` to install it after been uploaded or `apply package` to upload it and install it.

### `exec`

The exec action or command applies only to the clusters and packages.

#### Execute a command or script file on a `cluster`

Exec cluster is used to execute a remote command in all the nodes of the cluster or some of them specified with the flag `--nodes`, `-n`, or `--pool`, `-p`.

```bash
kubekit exec [cluster] NAME \
  --cmd command \
  --file script-file \
  --nodes node[,node ...] \
  --pools pool[,pool ...] \
  --sudo \
  --output json|yaml|toml \
  --pp
```

The `--sudo` flag is to allow KubeKit to use the `sudo` command to execute the command or script.

The output, error from StdErr and exit status will be printed on screen for every host the command was executed. This output could be on JSON, YAML or Toml, as specified by the flag `--output`.

#### Execute or install a previously copied `package`

```bash
kubekit exec package CLUSTER-NAME \
  --package-file FILE \
  --force-pkg
```

Use this command to install a previously copied package with the `copy package` command. 

The flag `--package-file` if to provide the file name of the package if the it is not the default one (`kubekit`). 

Some platforms or scenarios requires to use the `--force-pkg`, this flag is to force the installation of the package even on platforms that a package does not make sense such as EKS and AKS.

### `login`

The login command is used on clusters and a node. It's a required command for clusters as it provide the credentials to login or use the cluster platform. Each platform will have different kind of credentials. When used with a node, it's for login or ssh into a node.

#### Login or enter credentials of a `cluster`

```bash
kubekit login [cluster] NAME \
  --list \
  --access_key aws_access_key_id \
  --secret_key aws_secret_access_key \
  --region aws_default_region \
  --profile aws_profile \
  --server server_ip_or_dns \
  --username username \
  --password password
```

Login into a cluster is to provide the credentials to the platform where KubeKit is going to create or modify the Kubernetes cluster. It's a command that is required before execute `kubekit apply` and usually done after `kubekit init`.

Depending of the platform are the flags to provide:

- For **vSphere**, the required flags are: `--server`, `--username` and `--password`. The other flags are not required.

- For **vSphere** the variables are: **VSPHERE_SERVER**, **VSPHERE_USERNAME** and **VSPHERE_PASSWORD**

Example:

```bash
export VSPHERE_SERVER=153.0.0.101
export VSPHERE_USERNAME='username@vsphere.local'
export VSPHERE_PASSWORD='SuperSecure!Pa55w0rd'
```

If you are not comfortable with providing the credentials using the command flags, then don't do it and KubeKit will ask for the credentials. The input won't be print out into the console.

Use the flag `--list` to show the credentials. When it's sensitive information such as passwords or keys, KubeKit will show only stars `*` or stars and the last 4 characters.

There is a 3rd way to provide the credentials which is through environment variables. Using the following environment variables, KubeKit will use them as credentials, and the `--list` flag will show them [partially or completely hide] if they are set.


- For **EC2** or **EKS**, There are 3 options for authentication. The order of precedence is: cli options, aws profile, env variables:
  1. `--access_key`, `--secret_key`, `--region`
  2. `--profile`
  3. using environment variables.
  * `--session_token` is optional if you require aws session token authentication.

 AWS profile information can be found here: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html


 Session token information can be found here: See (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html) "Using Temporary Security Credentials with the AWS CLI" for more details.

For **EC2** and **EKS** the variables are: **AWS_ACCESS_KEY_ID**, **AWS_SECRET_ACCESS_KEY**, **AWS_DEFAULT_REGION** OR **AWS_PROFILE**
* If you require session tokens the additional variable:  **AWS_SESSION_TOKEN** will be used as well.

Example:

```bash
export AWS_ACCESS_KEY_ID='YOUR_AWS_ACCESS_KEY'
export AWS_SECRET_ACCESS_KEY='YOUR_AWS_SECRET_KEY'
export AWS_DEFAULT_REGION=us-west-2
```

#### Login to a `node`

The login to a node is not like login to a cluster, it's actually login into a node using SSH.

```bash
kubekit login node node-name-ip-or-dns \
  --cluster cluster-name
```

It's only possible to login into a single node. Use the node name, IP address or DNS to indicate the node to login in.

It's required to enter the cluster name where this node belongs as KubeKit will use the username and SSL keys for this cluster.

### `describe`

The describe command is similar to get but will give more information, it's limited to a number of objects and the output is in the formats JSON, YAML or TOML, by default is JSON. The objects you can get a description are: clusters, template, nodes or package.

The `--pp` flag is used only when the output is JSON and will display the JSON in a pretty print format or indented.

#### Describe `clusters`

```bash
kubekit describe [clusters] NAME[,NAME ...] \
  --output json|yaml|toml \
  --pp
```

#### Describe `templates`

```bash
kubekit describe templates NAME[,NAME ...]
  --output json|yaml|toml \
  --pp
```

#### Describe `nodes`

```bash
kubekit describe nodes node-name-ip-or-dns \
  --cluster cluster-name \
  --output json|yaml|toml \
  --pp
```

#### Describe `packages`

```bash
kubekit describe packages CLUSTER-NAME[,CLUSTER-NAME ...] \
  --package-files file,[file ...]
```

Display the content or manifest of the package file (RPM or DEB) located in the cluster directory of the given cluster names. If there are packages in a different location, use the `--package-files` or `-f` to indicate the different package files to describe.

### `start`, `stop` and `restart`

The start, stop and restart commands applies to server, clusters and nodes. As the name implies they are to start, stop or restart a KubeKit server, a cluster, a single or multiple nodes of a cluster filtered by node name, IP, DNS or by the pool name.

#### Start/Stop `cluster`

```bash
kubekit start|stop|restart [cluster] NAME[,NAME ...] \
  --nodes node[,node ...] \
  --pools pool[,pool ...]
```

To start, stop or restart the entire cluster just use the cluster name, no other flag is required, and you can give a list of clusters to restart.

If you only want to restart a node or set of nodes, use the flags `--nodes` or `-n` to list the nodes or use wildcards (`*`, `?`) or ranges (`[2-5]`). You may also restart/start/stop all the nodes in a pool or from multiple pools using the flags `--pool` or `-p`. With pools you can also use wildcards and ranges.

#### Start/Stop `server`

The command for server have different parameters:

```bash
kubekit start|stop|restart server \
  --host 0.0.0.0 \
  --port 5823 \
  --grpc-port 0 \
  --no-http \
  --healthz-port 5823 \
  --cert-dir $HUBEKIT_HOME/server/pki \
  --tls-cert-file /path/to/my/certs/kubekit.cert \
  --tls-private-key-file /path/to/my/certs/kubekit.key \
  --ca-file /path/to/my/certs/ca.key \
  --insecure
```

The `--host` parameter is the IP address or hostname for KubeKit Server to serve on, the default value is `0.0.0.0`. The default port where KubeKit serve is on, is `5823` but if you want to use a different port use the flag `—-port` to define it.

The server by default expose the API as a REST/HTTP API and as a gRPC API on same port (`--port`), or different ports if `--grpc-port` is set (REST on `--port` and gRPC on `--grpc-port`). However, it is possible to not expose the REST/HTTP API with the flag `--no-http`. It is not possible to not expose the gRPC API, this is always exposed, it's up to you to use it or not.

The gRPC API is recommended for any kind of access to KubeKit Server: from internal applications/services running on the same cluster/network, and from external applications or services. The REST/HTTP API is recommended for external applications or services or if gRPC is not supported the client.

The parameter  `--healthz-port` is used to define the port for the healthz server to serve on, the default value is the same as the server port  `5823`. Set the port to `0` to disable the healthz server.

The flag `--cert-dir` is the directory where the TLS certs are located. If `--tls-cert-file` and `--tls-private-key-file` are provided, this flag will be ignored. The default certificates directory is the directory `/server/pki` in the KubeKit directory which by default is `~/kubekit.d`. The flag `--tls-cert-file` is the file containing x509 Certificate used for serving HTTPS and `--tls-private-key-file` is the file containing x509 private key matching `--tls-cert-file`.

If `--tls-cert-file` and `--tls-private-key-file` are not provided, and there isn't any certificate/key in the `--cert-dir` directory, a self-signed certificate and key are generated for the public address and saved to the directory passed to `--cert-dir`.

If `--ca-file` is set, the given TLS certificate and private key should match with it. If no TLS certificate nor private key is set then they will be generated using this CA certificate file. A client certificate will be generated with this CA certificate and any request presenting a client certificate signed by one of the authorities in the ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.

The `--insecure` flag is set when TLS is not required. This may be used in non-production environments or if the access to KubeKit Server is only from internal applications and there is no exposure of KubeKit outside of the cluster or network.

**<u>Note</u>**: At this time, the only way to have gRPC and REST/HTTP API running on different ports is using the insecure mode.

<u>Example</u>:

```bash
kubekit start server
```

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

### `scale`

The scale command is used to increase or decrease the number of nodes of a clusters. It will modify the cluster configuration file and apply the changes, which is to increase/decrease the number of nodes and configure Kubernetes and services in the new nodes.

```bash
kubekit scale [cluster] cluster-name pool-name=[+|-]number[ pool-name=[+|-]number ...]
```

This command will set a new number of nodes in the given pool names of the given cluster. If the sign of plus `+` or minus `-` are used before the number, it means that will increase or decrease the existing number of nodes by that given number.

For example, if we have the cluster named `kkdemo` with one master (pool name is `master`) and 3 workers (pool name is `worker`). The following commands will scale to:

Executing `kubekit scale kkdemo master=2` scales to: masters = 2, workers = 3.

Executing `kubekit scale kkdemo master=2 worker=2` scales to: masters = 2, workers = 2.

To the previous cluster (2x2), executing`kubekit scale kkdemo master=-1 worker=+2` scales to: masters = 1, workers = 4.

The scale command will basically modify the number of nodes in the cluster configuration file and apply the changes like `kubekit apply` command would do. So, you may also scale the cluster that way, the `scale` command is just a shortcut.

## Implementation matrix

There is a total of **36 commands**, **19** of them are done, fully implemented and tested, **5** of them implemented but not fully tested, the rest **12** are in the backlog without estimate sprint or implementation date yet.

| Verb                | Noun             | Implemented | Tested     | Sprint |
| ------------------- | ---------------- | ----------- | ---------- | ------ |
| help                |                  | 100%        | 100%       | 30     |
| version             |                  | 100%        | 100%       | 30     |
| show-config         |                  | 100%        | 100%       | 30     |
| **completion**      |                  | **0%**      | **0%**     | *****  |
| init                | cluster          | 100%        | 100%       | 30     |
|                     | template         | 100%        | **50% **** | 30     |
|                     | certificates     | 100%        | **50% **** | 30     |
|                     | **package**      | **5%**      | **0%**     | *****  |
| apply               | cluster          | 100%        | 100%       | 30     |
|                     | package          | 100%        | **50% **** | 30     |
| delete              | cluster          | 100%        | 100%       | 30     |
|                     | cluster-config   | 100%        | 100%       | 30     |
|                     | **template**     | **5%**      | **0%**     | *****  |
|                     | **file**         | **5%**      | **0%**     | *****  |
|                     | **package**      | **0%**      | **0%**     | *****  |
| edit                | clusters         | 100%        | 100%       | 30     |
| get                 | clusters         | 100%        | 100%       | 30     |
|                     | nodes            | 100%        | 100%       | 31     |
|                     | **templates**    | **5%**      | **0%**     | *****  |
|                     | env              | 100%        | 100%       | 34     |
| copy                | **clusters**     | **5%**      | **0%**     | *****  |
|                     | cluster-config   | 100%        | 100%       | 31     |
|                     | **template**     | **5%**      | **0%**     | *****  |
|                     | files            | 100%        | 100%       | 30     |
|                     | **certificates** | **5%**      | **0%**     | *****  |
|                     | package          | 100%        | 100%       | 30     |
| exec                | cluster          | 100%        | **75% **** | 30     |
|                     | package          | 100%        | **50% **** | 30     |
| login               | cluster          | 100%        | 100%       | 30     |
|                     | node             | 100%        | 100%       | 31     |
| describe            | cluster          | 100%        | 100%       | 31     |
|                     | **template**     | **5%**      | **0%**     | *****  |
|                     | nodes            | 100%        | 100%       | 31     |
|                     | packages         | 5%          | 0%         | 31     |
| **[re]start, stop** | **cluster**      | **5%**      | **0%**     | *****  |
| **scale**           | **cluster**      | **5%**      | **0%**     | *****  |

(*****) Task to implement this command is in backlog (12 commands)

(******) Command implemented but not fully tested due to dependencies (5 commands)

(** **) Considering to remove the command (1 command)
