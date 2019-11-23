package eks

// Code generated automatically by 'go run codegen/main.go --pkg <pkg> --src <pkg>/templates --dst <pkg>/code.go'; DO NOT EDIT THIS FILE.

func init() {
	ResourceTemplates = map[string]string{
		"data-sources": dataSourcesTpl,
		"output":       outputTpl,
		"provider":     providerTpl,
		"resources":    resourcesTpl,
		"variables":    variablesTpl,
	}
}

// Expressions in the templates
/**
data-sources : {{ .AwsVpcID }}
data-sources : {{ .AwsVpcID }}
data-sources : {{ range $k, $v := .NodePools }}
data-sources : {{ if gt $v.Count 0 }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ $v.Count }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ Dash ( Lower $.ClusterName ) }}
data-sources : {{ Dash ( Lower $v.Name ) }}
data-sources : {{ end }}
data-sources : {{ end }}
data-sources : {{ with .Route53Name }}
data-sources : {{ range . }}
data-sources : {{ Dash ( Lower . ) }}
data-sources : {{ Dash ( Lower . ) }}
data-sources : {{ $.AwsVpcID }}
data-sources : {{ end }}
data-sources : {{ end }}
data-sources : {{ with .S3Buckets }}
data-sources : {{ range . }}
data-sources : {{ . }}
data-sources : {{ . }}
data-sources : {{ end }}
data-sources : {{ end }}
output : {{- range $k, $v := $.NodePools -}}
output : {{- range $i := Count $v.Count  }}
output : {{- Dash $v.Name }}
output : {{ $i }}
output : {{- Dash $v.Name }}
output : {{ $i }}
output : {{- Dash $v.Name }}
output : {{ $i }}
output : {{- Dash $v.Name }}
output : {{ $i }}
output : {{ $v.Name }}
output : {{ Dash ( Lower $k ) }}
output : {{ end }}
output : {{ end }}
output : {{- $first := true -}}
output : {{- range $k, $v := $.ElasticFileshares -}}
output : {{- if $first -}}
output : {{- $first = false -}}
output : {{- else -}}
output : {{end -}}
output : {{- Dash ( Lower $k ) }}
output : {{- Dash ( Lower $k ) }}
output : {{- Dash ( Lower $k ) }}
output : {{- end }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ .KubernetesVersion }}
resources : {{ QuoteList .ClusterLogsTypes }}
resources : {{ QuoteList .ClusterSecurityGroups }}
resources : {{ QuoteList .IngressSubnets }}
resources : {{ .EndpointPublicAccess }}
resources : {{ .EndpointPrivateAccess }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ with .S3Buckets }}
resources : {{ range . }}
resources : {{ . }}
resources : {{ . }}
resources : {{ . }}
resources : {{ . }}
resources : {{ end }}
resources : {{ end }}
resources : {{ with .Route53Name }}
resources : {{ range . }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower . ) }}
resources : {{ Dash ( Lower . ) }}
resources : {{ end }}
resources : {{ end }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{ Trim .PublicKey }}
resources : {{ range $k, $v := .NodePools }}
resources : {{ if $v.PGStrategy }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{$v.PGStrategy }}
resources : {{ end }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{- if $v.AwsAmi -}}
resources : {{- $v.AwsAmi -}}
resources : {{- else -}}
resources : {{- end }}
resources : {{ $v.AwsInstanceType }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ QuoteList $v.SecurityGroups }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ $v.RootVolumeSize -}}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ if $v.PGStrategy -}}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{- end }}
resources : {{ $v.Count }}
resources : {{ $v.Count }}
resources : {{ $v.Count }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ if $v.PGStrategy }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{- end }}
resources : {{ QuoteList $v.Subnets }}
resources : {{ Dash ( Lower $k ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower $v.Name ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ Dash ( Lower $.ClusterName ) }}
resources : {{ end }}
resources : {{ range $k, $v := .ElasticFileshares }}
resources : {{ Dash ( Lower $k  ) }}
resources : {{ Dash $.ClusterName }}
resources : {{ Dash $k }}
resources : {{ $v.PerformanceMode }}
resources : {{ $v.ThroughputMode }}
resources : {{ $v.Encrypted }}
resources : {{ Dash $.ClusterName }}
resources : {{ Dash ( Lower $k ) }}
resources : {{ range $s := AllSubNets }}
resources : {{ Dash $.ClusterName }}
resources : {{ Dash ( Lower $k ) }}
resources : {{ $s }}
resources : {{ Dash ( Lower $k ) }}
resources : {{ Dash ( Lower $k ) }}
resources : {{ $s }}
resources : {{ QuoteList AllSecGroups }}
resources : {{ end }}
resources : {{ end }}
variables : {{ range $k, $v := .NodePools }}
variables : {{ Dash ( Lower $v.Name ) }}
variables : {{ $.MaxMapCount }}
variables : {{- if IsFastEphemeral $v }}
variables : {{- end }}
variables : {{- if ne ( len $v.KubeletNodeLabels ) 0 -}}
variables : {{- Join $v.KubeletNodeLabels "," -}}
variables : {{- else -}}
variables : {{- Join $.DefaultNodePool.KubeletNodeLabels "," -}}
variables : {{- end -}}
variables : {{- if ne ( len $v.KubeletNodeTaints ) 0 -}}
variables : {{- Join $v.KubeletNodeTaints "," -}}
variables : {{- else -}}
variables : {{- Join $.DefaultNodePool.KubeletNodeTaints "," -}}
variables : {{- end -}}
variables : {{- if Contains $v.AwsInstanceType "large" -}}
variables : {{- else -}}
variables : {{- end -}}
variables : {{ $.ClusterName }}
variables : {{ end }}
**/

const dataSourcesTpl = `# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# data_sources.tf collects data and set's variables to be used later.  
# It does nothing to modify the images

data "aws_vpc" "vpc" {
  id = "{{ .AwsVpcID }}"
}

data "aws_subnet_ids" "vpc_subnets" {
  vpc_id = "{{ .AwsVpcID }}"
}

data "aws_region" "current" {}

data "aws_ami" "eks-node" {
  depends_on = ["aws_eks_cluster.kubekit"]
  filter {
    name   = "name"
    values = ["amazon-eks-node-${aws_eks_cluster.kubekit.version}-v*"]
  }

  most_recent = true
  owners      = ["602401143452"] # Amazon EKS AMI Account ID
}

{{ range $k, $v := .NodePools }}

  {{ if gt $v.Count 0 }}
data "aws_instance" "{{ Dash ( Lower $v.Name ) }}" {
  count = "{{ $v.Count }}"
  depends_on = ["data.aws_instances.{{ Dash ( Lower $v.Name ) }}"]
  instance_id = "${data.aws_instances.{{ Dash ( Lower $v.Name ) }}.ids[count.index]}"
}
  

data "aws_instances" "{{ Dash ( Lower $v.Name ) }}" {
  depends_on = [ "aws_autoscaling_group.node-{{ Dash ( Lower $v.Name ) }}",
  ]
  instance_tags = {
    Name = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" 
  }
}
  {{ end }}
{{ end }}

{{ with .Route53Name }}
	{{ range . }}
data "aws_route53_zone" "zone-{{ Dash ( Lower . ) }}" {
  name         = "{{ Dash ( Lower . ) }}"
  vpc_id = "{{ $.AwsVpcID }}"
}
	{{ end }} 
{{ end }}

{{ with .S3Buckets }}
	{{ range . }}
data "aws_s3_bucket" "{{ . }}" {
  bucket = "{{ . }}"
}
	{{ end }} 
{{ end }}
`

const outputTpl = `# Outputs
# ==============================================================================
output "endpoint" {
  value = "${aws_eks_cluster.kubekit.endpoint}"
}

output "certificate-authority" {
  value = "${aws_eks_cluster.kubekit.certificate_authority.0.data}"
}

output "role-arn" {
  value = "${aws_iam_role.cluster-node.arn}"
}

output "nodes" {
  value = [ {{- range $k, $v := $.NodePools -}} {{- range $i := Count $v.Count  }}
      "{\"private_ip\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.private_ip}\",\"public_ip\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.public_ip}\",\"public_dns\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.public_dns}\",\"private_dns\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.private_dns}\",\"pool\": \"{{ $v.Name }}\",\"role\": \"{{ Dash ( Lower $k ) }}\"}",{{ end }}{{ end }}
  ]
}

output "elastic-fileshares" {
  value = "[ {{- $first := true -}}{{- range $k, $v := $.ElasticFileshares -}}
    {{- if $first -}}{{- $first = false -}}{{- else -}}, {{end -}}
    { \"efs_name\": \"{{- Dash ( Lower $k ) }}\",\"efs_id\": \"${aws_efs_file_system.efs-
    {{- Dash ( Lower $k ) }}.id}\",\"efs_dns\": \"${aws_efs_file_system.efs-
    {{- Dash ( Lower $k ) }}.dns_name}\",\"efs_region\": \"${data.aws_region.current.name}\" }
    {{- end }} ]"
}`

const providerTpl = `# Provider 
# ==============================================================================
provider "aws" {
  access_key = "${ var.access_key }"
  secret_key = "${ var.secret_key }"
  region     = "${ var.region }"
  token      = "${ var.token }"
}
`

const resourcesTpl = `# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# resources.tf collects creates the resources that will be used with the image.  
# Be careful with what you create as a resource, as you can overwrite existing 
# infrastructure easily.

# EKS Control Plane
# ==============================================================================
resource "aws_iam_role" "cluster" {
  name = "{{ Dash ( Lower .ClusterName ) }}-iam-role"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  depends_on = ["aws_iam_role.cluster"]
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = "{{ Dash ( Lower .ClusterName ) }}-iam-role"
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSServicePolicy" {
  depends_on = ["aws_iam_role.cluster"]
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = "{{ Dash ( Lower .ClusterName ) }}-iam-role"
}

resource "aws_eks_cluster" "kubekit" {
  name     = "{{ Dash ( Lower .ClusterName ) }}"
  role_arn = "${aws_iam_role.cluster.arn}"
  version = "{{ .KubernetesVersion }}"
  enabled_cluster_log_types = [ {{ QuoteList .ClusterLogsTypes }} ]

  vpc_config {
    security_group_ids = [ {{ QuoteList .ClusterSecurityGroups }} ]
    subnet_ids = [ {{ QuoteList .IngressSubnets }} ]
    endpoint_public_access = "{{ .EndpointPublicAccess }}"
    endpoint_private_access = "{{ .EndpointPrivateAccess }}"
  }

  depends_on = [
    "aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy",
    "aws_iam_role_policy_attachment.cluster-AmazonEKSServicePolicy",
  ]
}

# EKS Nodes
# ==============================================================================
resource "aws_iam_role" "cluster-node" {
  name = "{{ Dash ( Lower .ClusterName ) }}-node"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

{{ with .S3Buckets }}
	{{ range . }}
resource "aws_iam_role_policy" "{{ . }}-policy" {
  depends_on = ["aws_iam_role.cluster-node"]
  name  = "{{ . }}-policy"
  role  = "${aws_iam_role.cluster-node.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
              "s3:PutObject",
              "s3:GetObject",
              "s3:DeleteObject",
              "s3:ListBucket"
            ],
            "Resource": "${data.aws_s3_bucket.{{ . }}.arn}/*"
        },
        {
            "Effect": "Allow",
            "Action": [
              "s3:ListBucket",
              "s3:GetBucketLocation"
            ],
            "Resource": "${data.aws_s3_bucket.{{ . }}.arn}"
        },
        {
            "Effect": "Allow",
            "Action": [
              "s3:ListAllMyBuckets"
            ],
            "Resource": "arn:aws:s3:::*"
        }
    ]
}
EOF
}
	{{ end }} 
{{ end }}

{{ with .Route53Name }}
	{{ range . }}
resource "aws_iam_role_policy" "route53-policy" {
  depends_on = ["aws_iam_role.cluster-node"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-zone-{{ Dash ( Lower . ) }}-route53-policy"
  role       = "${aws_iam_role.cluster-node.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action":[
        "route53:ChangeResourceRecordSets",
        "route53:ListResourceRecordSets"
      ],
      "Resource":"arn:aws:route53:::hostedzone/${data.aws_route53_zone.zone-{{ Dash ( Lower . ) }}.id}"
    }
  ]
}
EOF
}
  {{ end }}
{{ end }}

resource "aws_iam_role_policy_attachment" "node-AmazonEKSWorkerNodePolicy" {
  depends_on = ["aws_iam_role.cluster-node"]
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = "{{ Dash ( Lower .ClusterName ) }}-node"
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKS-CNI-Policy" {
  depends_on = ["aws_iam_role.cluster-node"]
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = "{{ Dash ( Lower .ClusterName ) }}-node"
}

resource "aws_iam_role_policy_attachment" "node-AmazonEC2ContainerRegistryReadOnly" {
  depends_on = ["aws_iam_role.cluster-node"]
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = "{{ Dash ( Lower .ClusterName ) }}-node"
}

resource "aws_iam_instance_profile" "node" {
  depends_on = ["aws_iam_role.cluster-node"]
  name = "{{ Dash ( Lower .ClusterName ) }}-node-iam-profile"
  role = "{{ Dash ( Lower .ClusterName ) }}-node"
}

resource "aws_key_pair" "keypair" {
  // TODO need to verify if key name change will cause destruction on existing 1.0 systems
  key_name   = "{{ Dash ( Lower .ClusterName ) }}-key"
  public_key = "{{ Trim .PublicKey }}"
}

{{ range $k, $v := .NodePools }}
  {{ if $v.PGStrategy }}
resource "aws_placement_group" "node-pool-{{ Dash ( Lower $v.Name ) }}" {
  name     = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}"

  strategy = "{{$v.PGStrategy }}"
  
}
  {{ end }}

resource "aws_launch_configuration" "node-{{ Dash ( Lower $v.Name ) }}" {
  associate_public_ip_address = true
  ebs_optimized               = true
  iam_instance_profile        = "{{ Dash ( Lower $.ClusterName ) }}-node-iam-profile"
  image_id                    = "
  {{- if $v.AwsAmi -}}
    {{- $v.AwsAmi -}}
  {{- else -}}
    ${data.aws_ami.eks-node.id}
  {{- end }}"
  instance_type               = "{{ $v.AwsInstanceType }}"
  name_prefix                 = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}-"
  security_groups             = [{{ QuoteList $v.SecurityGroups }}]
  user_data_base64            = "${base64encode(local.node-{{ Dash ( Lower $v.Name ) }}-userdata)}"
  key_name                    = "{{ Dash ( Lower $.ClusterName ) }}-key"

  root_block_device {
    delete_on_termination = true
    volume_size           = "{{ $v.RootVolumeSize -}}"
    volume_type           = "gp2"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "node-{{ Dash ( Lower $v.Name ) }}" {
  {{ if $v.PGStrategy -}}
  depends_on           = [ "aws_placement_group.node-pool-{{ Dash ( Lower $v.Name ) }}" ]
  {{- end }}
  desired_capacity     = "{{ $v.Count }}"
  max_size             = "{{ $v.Count }}"
  min_size             = "{{ $v.Count }}"
  launch_configuration = "${aws_launch_configuration.node-{{ Dash ( Lower $v.Name ) }}.name}"
  {{ if $v.PGStrategy }}
  placement_group      = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}"
  {{- end }}
  vpc_zone_identifier  = [ {{ QuoteList $v.Subnets }} ]
  tag {
    key                 = "NodePool"
    value               = "{{ Dash ( Lower $k ) }}"
    propagate_at_launch = true
  }

  tag {
    key                 = "Name"
    value               = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" // needs to be just "-worker" on 1.0
    propagate_at_launch = true
  }
 
  tag {
    key                 = "ClusterName"
    value               = "{{ Dash ( Lower $.ClusterName ) }}"
    propagate_at_launch = true
  }

  tag {
    key                 = "Project"
    value               = "KubeKit"
    propagate_at_launch = true
  }

  tag {
    key                 = "kubernetes.io/cluster/{{ Dash ( Lower $.ClusterName ) }}"
    value               = "owned"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
  }
}
{{ end }}


{{ range $k, $v := .ElasticFileshares }}
resource "aws_efs_file_system" "efs-{{ Dash ( Lower $k  ) }}" {
  creation_token = "{{ Dash $.ClusterName }}-efs-{{ Dash $k }}"
  
  performance_mode = "{{ $v.PerformanceMode }}"
  throughput_mode = "{{ $v.ThroughputMode }}"
  encrypted = "{{ $v.Encrypted }}"

  tags = {
    Name = "{{ Dash $.ClusterName }}-efs-{{ Dash ( Lower $k ) }}"
  }
}

  {{ range $s := AllSubNets }}
resource "aws_efs_mount_target" "{{ Dash $.ClusterName }}-efs-{{ Dash ( Lower $k ) }}-{{ $s }}-mount" {
  depends_on = ["aws_efs_file_system.efs-{{ Dash ( Lower $k ) }}"]
  file_system_id = "${aws_efs_file_system.efs-{{ Dash ( Lower $k ) }}.id}"
  subnet_id      = "{{ $s }}"
  security_groups = [ {{ QuoteList AllSecGroups }} ]
}
  {{ end }}
{{ end }}`

const variablesTpl = `# Variables
# ==============================================================================

variable "access_key" {}
variable "secret_key" {}
variable "region"   {}
variable "token"   {}

locals {

{{ range $k, $v := .NodePools }}
  node-{{ Dash ( Lower $v.Name ) }}-userdata =  <<USERDATA
#!/bin/bash
ulimit -n 65535
cat <<EOF > /etc/security/limits.d/30-nofile.conf
root       soft    nofile     65536
root       hard    nofile     65536
EOF
cat <<EOF > /etc/sysctl.d/99-kubelet-network.conf
# Have a larger connection range available
net.ipv4.ip_local_port_range=1024 65000

# Reuse closed sockets faster
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_fin_timeout=15

# The maximum number of "backlogged sockets".  Default is 128.
net.core.somaxconn=4096
net.core.netdev_max_backlog=4096

# 16MB per socket - which sounds like a lot,
# but will virtually never consume that much.
net.core.rmem_max=16777216
net.core.wmem_max=16777216

# Various network tunables
net.ipv4.tcp_max_syn_backlog=20480
net.ipv4.tcp_max_tw_buckets=400000
net.ipv4.tcp_no_metrics_save=1
net.ipv4.tcp_rmem=4096 87380 16777216
net.ipv4.tcp_syn_retries=2
net.ipv4.tcp_synack_retries=2
net.ipv4.tcp_wmem=4096 65536 16777216
#vm.min_free_kbytes=65536

# Connection tracking to prevent dropped connections (usually issue on LBs)
net.netfilter.nf_conntrack_max=262144
net.ipv4.netfilter.ip_conntrack_generic_timeout=120
net.netfilter.nf_conntrack_tcp_timeout_established=86400

# ARP cache settings for a highly loaded docker swarm
net.ipv4.neigh.default.gc_thresh1=8096
net.ipv4.neigh.default.gc_thresh2=12288
net.ipv4.neigh.default.gc_thresh3=16384

# disable ipv6
net.ipv6.conf.all.disable_ipv6=1
net.ipv6.conf.default.disable_ipv6=1
net.ipv6.conf.lo.disable_ipv6=1

# set max_map_count for elasticsearch usage
vm.max_map_count={{ $.MaxMapCount }}

fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=524288
EOF
systemctl restart systemd-sysctl.service
set -o xtrace
systemctl stop kubelet
systemctl stop docker
  {{- if IsFastEphemeral $v }}
yum install rsync -y
pvcreate /dev/nvme[1-9]*n*
vgcreate vgdata /dev/nvme[1-9]*n*
lvcreate -l 100%FREE --type striped -n lvoldata vgdata
mkfs.xfs /dev/vgdata/lvoldata
mkdir -p /data
mount /dev/vgdata/lvoldata /data
echo "/dev/vgdata/lvoldata /data xfs defaults 0 0"  >> /etc/fstab
for directory in /var/lib/docker /var/lib/kubelet; do
  if [ ! -d "$${directory}" ]; then
    mkdir -p $${directory}
  fi
  if [ ! -z "$(ls -A $${directory})" ]; then
    rsync -avzh $${directory}/ /data/$(basename $${directory})
  fi
  rm -Rf $${directory}/*
  mount --bind /data/$(basename $${directory}) $${directory}
  echo "$${directory} /data/$(basename $${directory}) bind bind 0 0"  >> /etc/fstab
done
  {{- end }}
DOCKER_DAEMON=$(jq 'del(."default-ulimits")' /etc/docker/daemon.json)
echo "$DOCKER_DAEMON" > /etc/docker/daemon.json
systemctl start docker
/etc/eks/bootstrap.sh --kubelet-extra-args '--node-labels="
{{- if ne ( len $v.KubeletNodeLabels ) 0 -}}{{- Join $v.KubeletNodeLabels "," -}}
{{- else -}}{{- Join $.DefaultNodePool.KubeletNodeLabels "," -}}
{{- end -}}" --register-with-taints="
{{- if ne ( len $v.KubeletNodeTaints ) 0 -}}{{- Join $v.KubeletNodeTaints "," -}}
{{- else -}}{{- Join $.DefaultNodePool.KubeletNodeTaints "," -}}
{{- end -}}" --kube-reserved cpu=250m,memory=
{{- if Contains $v.AwsInstanceType "large" -}}1{{- else -}}0.5{{- end -}}
Gi,ephemeral-storage=1Gi --system-reserved cpu=250m,memory=0.2Gi,ephemeral-storage=1Gi --eviction-hard memory.available<0.5Gi,nodefs.available<10%' --apiserver-endpoint '${aws_eks_cluster.kubekit.endpoint}' --b64-cluster-ca '${aws_eks_cluster.kubekit.certificate_authority.0.data}' '{{ $.ClusterName }}'
USERDATA

{{ end }}
}
`