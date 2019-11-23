# KNOWN ISSUES
- This file is for documenting non Kubekit Known Issues: Kubernetes, etcd, docker, K8 + Operating Systems, etc.

## Generic linux issues
- You cannot specify more than 6 search domains in /etc/resolv.conf: https://access.redhat.com/solutions/58028 . Also see next line.
- Kubernetes requires 1 nameserver and 1 search entry to be available, otherwise there will be issues. See https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/#known-issues for more details.

## Platform Issues
- ONLY VSphere works in automation via env variables.  AWS and OPENSTACK have issues.
- EKS pre-allocates a lot of IP addresses that may overwhelm your subnet CIDR.
```
Today, the EKS CNI plugin creates a “warm pool” of IP addresses by pre-allocating IP addresses on EKS nodes to reduce scheduling latency. In other words: because the instance already has IP addresses allocated to it, Kubernetes doesn’t need to wait for an IP address to be assigned before it can schedule a pod. However, there are some tradeoffs in this approach: if your EKS nodes are larger instance types and can support larger numbers of IP addresses, you might find that your nodes are hogging more IP addresses than you want.

You can use the WARM_IP_TARGET environment variable to tune the size of the IP address “warm pool.” You can define a threshold for available IP addresses below which L-IPAMD creates and attaches a new ENI to a node, allocates new IP addresses, and then adds them to the warm pool. This threshold can be configured using the WARM_IP_TARGET environment variable; it can also be configured in amazon-vpc-cni.yaml.

For example, an m4.4xlarge node can have up to 8 ENIs, and each ENI can have up to 30 IP addresses. This means that the m4.4xlarge could reserve up to 240 IP addresses from your VPC CIDR for its warm pool, even if there are no pods scheduled. Changing the WARM_IP_TARGET to a lower number will reduce how many IPs the node has attached, but if your number of pods scheduled exceeds the WARM_IP_TARGET, additional pod launches will require an EC2 AssignPrivateIpAddresses() API call, which can add latency to your pod startup times.

This parameter allows you to perform a balancing act. We recommend tuning it based on your pod launch needs: how many pods do you need to schedule and how fast do you need them to start up, versus how much of your VPC IP space you’d like your EKS nodes to occupy.
```
https://aws.amazon.com/blogs/opensource/vpc-cni-plugin-v1-1-available/

Since AWS deploys the latest versions of their VPC CNI when spinning up an EKS cluster, it is preferable to patch their daemonset.
To set the WARM_IP_TARGET mentioned above, run the following on a host that has access to the cluster:
```
export WARM_IP_TARGET=3  # change to however many works for you
kubectl -n=kube-system patch ds aws-node --patch "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"aws-node\",\"env\":[{\"name\":\"WARM_IP_TARGET\",\"value\":\"$WARM_IP_TARGET\"}]}]}}}}"
```
You can read more about it at the following links:
* https://github.com/aws/amazon-vpc-cni-k8s#eni-allocation
* https://github.com/aws/amazon-vpc-cni-k8s#cni-configuration-variables


## Breaking Changes

### Self-hosted control plane.

KK 2.0 use a self-hosted control plane - kubernetes runs in containers. This provides numerous advantages but changes some aspects of the deployment.

To use the current version of KK with the VSphere and AMI SUSE 12.SP3 images, the kubekit images rpm
MUST BE INJECTED prior to the "kubekit apply" command that applies the configuration. This provides the correct images for the self-hosted control plane and is applicable to the current 12.sp3 OVA and AMI images.

A typical work flow is, using "vsphere" as the platform and "clustername" as the name of the cluster. (Don't use "clustername" - that's stupid. Use something descriptive to you.)

```
# kubekit init -p vsphere clustername
# kubekit apply -p clustername
# kubekit copy package clustername -f kubekit-2.0.0.rpm
# kubekit exec clustername --cmd 'sudo rpm -Uvh /tmp/kubekit-2.0.0.rpm'
# kubekit apply clustername
```

The kubekit-2.0.0.rpm comes from the tgz download from the shared-service kubekit repository.

The "copy package" pushes the kubekit-2.0.0.rpm to /tmp on all the nodes in the cluster.

The "kubekit exec" installs the rpm. This MUST BE DONE in order for the cluster to work.
