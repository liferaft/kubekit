# KubeKit Client

KubeKit Server expose a gRPC and REST/HTTP API so any client in any programming language can be used to communicate with it. The KubeKit Client or `kubekitctl` was made to communicate with the server using the same commands used by KubeKit CLI.

KubeKit (`kubekit`) still can be used as a single binary with no dependencies to create, manage and terminate clusters. Just when it's use as a server in your host or a different host, is when KubeKit Client (`kubekitctl`) is used.

For more information about the KubeKit Client CLI refer to the CLI-UX document for it. Here I will show a quick example about how to use it.

In terminal #1 start the server, it will run with te default settings: secure and exposing gRPC and REST/HTTP in same port. Use `--debug` to see more information.

```bash
kubekit start server # --debug
```

In the following examples we'll use `jq` to make the JSON output pretty. Future version of the client will print the output in human readable format, JSON or YAML.

In terminal #2 let's execute the following commands:

1. [Check the KubeKit version](#1-check-the-kubekit-version)
2. [Initialize a cluster](#2-initialize-a-cluster)
3. [Update cluster parameters](#3-update-cluster-parameters)
4. [List the existing clusters](#4-list-the-existing-clusters)
5. [Verify the cluster configuration](#5-verify-the-cluster-configuration)
6. [Create the cluster](#6-create-the-cluster)
7. [Check the cluster status](#7-check-the-cluster-status)
8. [Access the cluster](#8-access-the-cluster)
9. [Terminate the cluster](#9-terminate-the-cluster)
10. [Delete clusters configuration and future access to them](#10-delete-clusters-configuration-and-future-access-to-them)
11. [Other commands](#11-other-commands)

### 1) Check the KubeKit version

```bash
kubekitctl version | jq
```

Just like `kubekit` this will show the version of the KubeKit server, the API version, the Kubernetes version used to create clusters as well as the Docker and etcd version.

### 2) Initialize a cluster

Just like `kubekit init` it's used to create the cluster configuration. As with `kubekit`, the client `kubekitctl` also accept parameters with the flag `--var`, actually it's the only way to pass data to the cluster configuration because the configuration file is not local. It also accepts the credentials with the different credentials flags for the following platforms:

* AWS and EKS: `--access_key`, `--secret_key`,  `--region`, `--session_token` and `--profile`

* Azure/AKS: `--subscription_id`, `--tenant_id`, `--client_id` and `--client_secret`

* vShpere and OpenStack: `--server`, `--username` and `--password`

Examples:

On EKS

```bash
export AWS_ACCESS_KEY_ID=....
export AWS_SECRET_ACCESS_KEY=...
export AWS_DEFAULT_REGION=us-west-2
echo "$AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY $AWS_DEFAULT_REGION"

kubekitctl init --platform eks \
    --access_key=$AWS_ACCESS_KEY_ID \
    --secret_key=$AWS_SECRET_ACCESS_KEY \
    --region=$AWS_DEFAULT_REGION \
    --var "aws_vpc_id=vpc-8d56b9e9" \
    --var "ingress_subnets=[subnet-5bddc82c,subnet-478a4123]" \
    --var "cluster_security_groups=[sg-502d9a37]" \
    --var "default_node_pool__worker_pool_subnets=[subnet-5bddc82c]" \
    --var "default_node_pool__security_groups=[sg-502d9a37]" \
    eks01
```

On vSphere

```bash
export VSPHERE_SERVER=....
export VSPHERE_USERNAME=...
export VSPHERE_PASSWORD=...
echo "$VSPHERE_SERVER $VSPHERE_USERNAME $VSPHERE_PASSWORD"

kubekitctl init cluster --platform vsphere \
 --server "$VSPHERE_SERVER" \
 --username "$VSPHERE_USERNAME" \
 --password "$VSPHERE_PASSWORD" \
 --var datacenter=Vagrant \
 --var datastore="sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03" \
 --var resource_pool="sd_vgrnt_01/Resources/vagrant01" \
 --var vsphere_net="dvpg_vm_550" \
 --var folder="machine/ja186051" \
 vs01
```

### 3) Update cluster parameters

If some parameter is missing or incorrect, the way to modify it is using the `update` subcommand.

Example:

```bash
kubekitctl update vs01 \
 --var folder="Discovered virtual machine/ja186051" \
 --var default_node_pool__template_name=Templates/LinkedClones/vmware-kubekit-os-01.45-19.08.05-200G
```

Just like `init`, use the flag `--var` and the credential flags to pass the new values.

To know the existing cluster configuration and the variables, use the subcommand `describe` shown below at step #5.

### 4) List the existing clusters

Use the `get cluster` command to list all the clusters and basic information about them such as name, platform, nodes and status.

```bash
kubekitctl get cluster | jq
```

Use the flag `--quiet` to list only the clusters name:

```bash
kubekitctl get cluster --quiet | jq
```

The `--quiet` flag is useful with other commands for example, if you'd like to delete the cluster configuration for all your existing clusters. Read below to see an example.

Just like in the CLI, you can use the `--filter` flag to filter the list of clusters to get. The parameters to filter by are:

- `name`: A cluster name. Use it when you want to know about one specific cluster.
- `nodes`: Number of nodes. Use it to know clusters with specific number of nodes, such as `0` or `3`.
- `platform`: Platform where the cluster is set. Use it to know the clusters on an specific platform.
- `status`: Cluster status. Values could be RUNNING, PROVISIONED, CONFIGURED, etc...
- `version`: Cluster version. 
- `path`: Cluster configuration files location/path.
- `url` or  `entrypoint`: Kubernetes entrypoint or HTTP address.
- `kubeconfig`: Kubeconfig path.

You can use multiple filters but not the same filter for the same parameter. For example, this filter `--filter platform=vsphere --filter platform=eks` will return the EKS or vSphere clusters but not both, so it does not work as expected.

In future versions the parameters may accept wildcards (i.e. filter by names that contain the text "public"), comparison operators (i.e. filter by number of nodes greater than 5) or negatives/not (i.e. filter by status not RUNNING).

Examples:

```bash
kubekitctl get cluster --filter name=eks01 | jq
kubekitctl get cluster --filter platform=vsphere --filter status=running | jq
kubekitctl get cluster --filter nodes=0 | jq
```

### 5) Verify the cluster configuration

To verify the cluster get the cluster information using `describe` sub-command. This command shows the basic information about a cluster such as name, platform, number of nodes and status.

```bash
kubekitctl describe eks01 | jq
```

The `--full` flag will show all the cluster information.

```bash
kubekitctl describe eks01 --full | jq
```

If you don't want to view all the cluster information, then use some of the following flags:

`--show-config`: Shows the cluster configuration, this includes the platform and Kubernetes configuration and resources to load.

`--show-nodes`: Prints the cluster nodes and basic information about them.

`--show-entrypoint`: Shows the entrypoint or HTTP address to access the cluster.

`--show-kubeconfig`: Prints the content of the Kubeconfig file required to access the cluster.

Example:

```bash
kubekitctl describe vs01 --show-config --show-nodes | jq
```

```bash
kubekitctl describe eks01 --show-entrypoint --show-kubeconfig | jq
```

With every flag the basic information will always be printed.

### 6) Create the cluster

Just like KubeKit CLI, the `apply` subcommand creates the Kubernetes cluster on the cluster platform.

```bash
kubekitctl apply eks01
```

Applying multiple clusters at same time is possible

```bash
kubekitctl apply vs01
```

The client does not wait until the server is ready to return control to the user. If you would like to know the status of the cluster use the `describe` sub-command.

### 7) Check the cluster status

Use the `describe` sub-command as described above on step #5

```bash
kubekitctl describe eks01 | jq
```

When the status is `RUNNING`, the server is ready. If not, the status will show the current status of the cluster.

### 8) Access the cluster

When ready, use the command `describe` to know the cluster entrypoint:

```bash
kubekitctl describe eks01 --show-entrypoint
```

And use the `get env` sub-command with the `eval` command to get the Kubeconfig file and export an environment variable to access it:

```bash
kubekitctl get env vs01

eval $(kubekitctl get env vs01)
```

The `get env` command download the Kubeconfig file to the file `~/.kube/NAME.kconf`  where NAME is the cluster name. If you like a different location use the flag `--kubeconfig-file` or `-f` to specify the file path and filename. For example:

```bash
eval $(kubekitctl get env eks01 -f ./config)
```

### 9) Terminate the cluster

If the `--all` flag is used, it just not terminate the cluster, it also delete the cluster configuration.

Use the `--force` to avoid KubeKit confirm before terminate the cluster.

```bash
kubekitctl delete eks01 --all --force
kubekitctl delete vs01 --force
```

### 10) Delete clusters configuration and future access to them

Use the command `delete clusters-config` to delete the cluster configuration if you didn't use the `--all` flag with the `delete` sub-command. It's important to mention that once the cluster configuration is deleted it cannot be recovered, so make sure that:

1. the cluster is terminated. Otherwise you will lost access to the cluster
2. the cluster is no needed anymore, or you'll have to create it again from scratch.

```bash
kubekitctl delete clusters-config eks01 --force
```

Use the flag `--force` to avoid KubeKit to ask for confirmation.

If you previously used the `get env` command to get the Kubeconfig file, you can use the same command with the flag `--unset` or `-u` to unset the environment variable and get the command to delete the Kubeconfig file.

```bash
kubekitctl get env vs01 --unset

eval $(kubekitctl get env vs01 -u)
```

The first command (`get env` without `eval`) will print instructions to remove the Kubeconfig file in case you want to remove any future access to the cluster.

Let's say you have a lot of clusters in your system and delete them one by one is too slow. Use the `delete cluster-config --force` and the `get cluster --quiet` command to delete them all with one line:

```bash
kubekitctl delete clusters-config --force $(kubekitctl get cluster -q | jq -r '.|@tsv ')
```

At this time all the commands prints the output in JSON format but this will change. Because of that the previous command requires the `jq` command to remove the JSON format and print the clusters in a one line list.

### 11) Other commands

**Get the clusters nodes**

Use the sub command `get nodes` to list the nodes existing in the cluster and some basic information such as IP addresses, DNS addresses and the pool the node is in. 

```bash
kubekitctl get nodes vs01
```

It also accept the flags `--nodes` and `--pools` to filter the nodes by its IP address or DNS using `--nodes` or filter them by pool using the `--pools` flag.

```bash
kubekitctl get nodes vs01 --nodes vs01-master-01
kubekitctl get nodes vs01 --nodes 10.25.151.205
kubekitctl get nodes vs01 --pools master
kubekitctl get nodes vs01 --pools worker --pools master
kubekitctl get nodes vs01 --pools worker,master
```

You can also print just the nodes IP public address using the flag `--quiet` or `-q`:

```bash
kubekitctl get nodes vs01 -q 
kubekitctl get nodes vs01 -q --pools master,worker
```

**Check Server Health**

Using the health service of the server you can use the subcommand `healthz` to check the server health.

```
kubekitctl healthz
```

If the server is using a different port for the health service, use the `--healthz-port` flag to specify the port.

**EKS cluster token**

Same as `kubekit token` the subcommand `toke` for the client works to return the token required for EKS clusters in order to authenticate the user when you use `kubectl`.

The cluster name is specified with `--cluster-id` or `-i`. It also accept the flag `--role` or `-r` to assume an IAM Role ARN before signing this token.

Example:

```bash
kubekitctl token --cluster-id eks01
```

This command may be deprecated when it's not required to use the EKS cluster.



