# Troubleshooting

<!-- TOC -->

- [Troubleshooting](#troubleshooting)
  - [How to get the `kubeconfig` file path or set the KUBECONFIG environment variable](#how-to-get-the-kubeconfig-file-path-or-set-the-kubeconfig-environment-variable)
  - [Basic questions when the `desc` command is not working as expected](#basic-questions-when-the-desc-command-is-not-working-as-expected)

<!-- /TOC -->

## How to get the `kubeconfig` file path or set the KUBECONFIG environment variable

**Option #1 (recommended):**

Use command `eval "$(kubekit get env CLUSTER-NAME)` and the kubeconfig path is in environment variable `KUBECONFIG`

Example:

```bash
$ CLUSTER_NAME=kkdemo     # here goes the name of your cluster

$ kubekit get env $CLUSTER_NAME
export KUBECONFIG=/root/.kubekit.d/clusters/6779e17f-a86a-4dfb-7495-94a2128f354e/certificates/kubeconfig
# Run this command to configure your shell:
# eval "$(kubekit get env kkdemo)"

$ eval "$(kubekit get env $CLUSTER_NAME)"

$ echo $KUBECONFIG
/root/.kubekit.d/clusters/6779e17f-a86a-4dfb-7495-94a2128f354e/certificates/kubeconfig

$ kubectl get nodes
```

This is the recommended way to get the path for kubeconfig file, and it will be in the KUBECONFIG environment variable, so after the `eval` command you are ready to use the `kubectl` command because it uses this env variable to locate the cluster. So, give it a try and let us know.

 **Option #2:**

Get the kubeconfig path using the `desc` command : `$ kubekit desc kkdemo | grep kubeconfig | cut -f2 -d: | tr -d ' '`

```bash
$ CLUSTER_NAME=kkdemo     # here goes the name of your cluster

$ export KUBECONFIG=$(kubekit desc $CLUSTER_NAME | grep kubeconfig | cut -f2 -d: | tr -d ' ')

$ echo $KUBECONFIG
/root/.kubekit.d/clusters/6779e17f-a86a-4dfb-7495-94a2128f354e/certificates/kubeconfig
```

**Option #3:**

Build the kubeconfig path with the path of the cluster.

The cluster path can be obtained with: `kubekit desc kkdemo | grep path | cut -f2 -d: | tr -d ' '`, so if you append to this path the following path: `/certificates/kubeconfig` you can get the kubeconfig full path. Like this:

```bash
$ CLUSTER_NAME=kkdemo     # here goes the name of your cluster

$ export KUBECONFIG="$(kubekit desc $CLUSTER_NAME | grep path | cut -f2 -d: | tr -d ' ')/certificates/kubeconfig"

$ echo $KUBECONFIG
/root/.kubekit.d/clusters/6779e17f-a86a-4dfb-7495-94a2128f354e/certificates/kubeconfig

```

## Basic questions when the `desc` command is not working as expected

In order to know more about why `desc` is not showing up all the parameters, ask the following questions:

- Send me the script you are executing on Jenkins to get the kubeconfig path
- Send me the rest of the lines in the logs or Jenkins output
- Tell me how many clusters are you spinning in the same environment
- Execute the following commands before the `kubekit desc cluster ...`:

```bash
CLUSTER_NAME=kkdemo     # here goes the name of your cluster
P=$(kubekit desc $CLUSTER_NAME | grep path | cut -f2 -d: | tr -d ' ')
kubekit get clusters
ls -al $P/
ls -al $P/*
ls -al $P/*/*
```

- Execute the following commands after the `kubekit desc cluster ...`:

```bash
CLUSTER_NAME=kkdemo     # here goes the name of your cluster
kubekit get env $CLUSTER_NAME
kubekit desc cluster $CLUSTER_NAME | grep kubeconfig
kubekit version --verbose
kubekit show-config --pp
```
