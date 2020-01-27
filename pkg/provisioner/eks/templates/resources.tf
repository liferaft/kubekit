# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

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

resource "null_resource" "eks_wait_for_iam_user_propagation" {
  depends_on = [
    "aws_iam_role.cluster",
    "aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy",
    "aws_iam_role_policy_attachment.cluster-AmazonEKSServicePolicy",
  ]

  provisioner "local-exec" {
    command = "sleep 45"
  }
}

resource "aws_eks_cluster" "kubekit" {
  depends_on = [
    "null_resource.eks_wait_for_iam_user_propagation",
  ]

  name     = "{{ Dash ( Lower .ClusterName ) }}"
  role_arn = aws_iam_role.cluster.arn
  version = "{{ .KubernetesVersion }}"
  enabled_cluster_log_types = [ {{ QuoteList .ClusterLogsTypes }} ]

  vpc_config {
    security_group_ids = [ {{ QuoteList .ClusterSecurityGroups }} ]
    subnet_ids = [ {{ QuoteList .IngressSubnets }} ]
    endpoint_public_access = "{{ .EndpointPublicAccess }}"
    endpoint_private_access = "{{ .EndpointPrivateAccess }}"
  }
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

resource "aws_iam_role_policy" "fsx-policy" {
  depends_on = ["aws_iam_role.cluster-node"]
  name  = "{{ Dash ( Lower .ClusterName ) }}-fsx-policy"
  role  = aws_iam_role.cluster-node.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "iam:CreateServiceLinkedRole",
          "iam:AttachRolePolicy",
          "iam:PutRolePolicy"
          ],
         "Resource": "arn:aws:iam::*:role/aws-service-role/fsx.amazonaws.com/*"
      },
      {
        "Effect": "Allow",
        "Action": [
          "fsx:*"
        ],
        "Resource": ["*"]
      }
    ]
}
EOF
}

{{ with .S3Buckets }}
	{{ range . }}
resource "aws_iam_role_policy" "s3-{{ Dash ( Lower . ) }}-policy" {
  depends_on = ["aws_iam_role.cluster-node"]
  name  = "{{ Dash ( Lower $.ClusterName ) }}-s3-{{ Dash ( Lower . ) }}-policy"
  role  = aws_iam_role.cluster-node.id

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
            "Resource": "${data.aws_s3_bucket.{{ Dash ( Lower . ) }}.arn}/*"
        },
        {
            "Effect": "Allow",
            "Action": [
              "s3:ListBucket",
              "s3:GetBucketLocation"
            ],
            "Resource": "${data.aws_s3_bucket.{{ Dash ( Lower . ) }}.arn}"
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
resource "aws_iam_role_policy" "route53-policy-zone-{{ Dash ( Lower . ) }}" {
  depends_on = ["aws_iam_role.cluster-node"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-zone-{{ Dash ( Lower . ) }}-route53-policy"
  role       = aws_iam_role.cluster-node.id

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

resource "null_resource" "node_wait_for_iam_user_propagation" {
  depends_on = [
    "aws_iam_role.cluster-node",
    "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
    "aws_iam_role_policy_attachment.node-AmazonEKS-CNI-Policy",
    "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
    "aws_iam_instance_profile.node",
  ]

  provisioner "local-exec" {
    command = "sleep 45"
  }
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
  depends_on = [
    "null_resource.node_wait_for_iam_user_propagation",
  ]

  associate_public_ip_address = true
  ebs_optimized               = true
  iam_instance_profile        = "{{ Dash ( Lower $.ClusterName ) }}-node-iam-profile"
  image_id                    =
  {{- if $v.AwsAmi -}}
    "{{- $v.AwsAmi -}}"
  {{- else -}}
    data.aws_ami.eks-node.id
  {{- end }}
  instance_type               = "{{ $v.AwsInstanceType }}"
  name_prefix                 = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}-"
  security_groups             = [{{ QuoteList $v.SecurityGroups }}]
  user_data_base64            = base64encode(local.node-{{ Dash ( Lower $v.Name ) }}-userdata)
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
  launch_configuration = aws_launch_configuration.node-{{ Dash ( Lower $v.Name ) }}.name
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
  file_system_id = aws_efs_file_system.efs-{{ Dash ( Lower $k ) }}.id
  subnet_id      = "{{ $s }}"
  security_groups = [ {{ QuoteList AllSecGroups }} ]
}
  {{ end }}
{{ end }}