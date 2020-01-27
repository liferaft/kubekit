# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# resources.tf collects creates the resources that will be used with the image.  
# Be careful with what you create as a resource, as you can overwrite existing 
# infrastructure easily.

# AWS Instances
# ==============================================================================
{{ $masterNodePool := MasterPool $.NodePools }}

{{ range $k, $v := .NodePools }}
  {{ if and $v.PGStrategy (isPGStrategy $v.PGStrategy) }} 

resource "aws_placement_group" "{{ Dash ( Lower $.ClusterName ) }}-node-pool-{{ Dash ( Lower $v.Name ) }}" {
  name     = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}"

  strategy = "{{ $v.PGStrategy }}"
}
  {{- end }}

resource "aws_launch_configuration" "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" {
  depends_on = ["aws_iam_instance_profile.kube-{{ Dash ( Lower $v.Name )  }}-profile"]
  associate_public_ip_address = true
  ebs_optimized               = true
  iam_instance_profile        = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $v.Name ) }}-profile"
  image_id                    = "{{ $v.Ami }}"
  instance_type               = "{{ $v.InstanceType }}"
  name_prefix                 = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}-"
  security_groups             = [{{ QuoteList $v.SecurityGroups }}]
  user_data_base64            = base64encode(local.node-{{ Dash ( Lower $v.Name ) }}-userdata)
  key_name                    = "{{ Dash ( Lower $.ClusterName ) }}-key"

  root_block_device {
    delete_on_termination = true
    volume_size           = "{{ $v.RootVolumeSize -}}"
    volume_type           = "{{ $v.RootVolumeType -}}"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" {
  depends_on           = [ 
      {{ if and $v.PGStrategy (isPGStrategy $v.PGStrategy) }} 
    "aws_placement_group.{{ Dash ( Lower $.ClusterName ) }}-node-pool-{{ Dash ( Lower $v.Name ) }}",
      {{- end }}
    "aws_launch_configuration.{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" 
  ]
  name                 = "{{ Dash ( Lower $.ClusterName ) }}-node-pool-{{ Dash ( Lower $v.Name ) }}"
  desired_capacity     = "{{ $v.Count }}"
  max_size             = "{{ $v.Count }}"
  min_size             = "{{ $v.Count }}"
  launch_configuration = aws_launch_configuration.{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}.name
  {{ if $v.PGStrategy }}
  placement_group      = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}"
  {{- end }}
  vpc_zone_identifier  = [ {{ QuoteList $v.Subnets }} ]
  {{ if eq $v.Name $masterNodePool.Name -}}
  target_group_arns    = [
    aws_alb_target_group.tgt-grp-api.arn,
    aws_alb_target_group.tgt-grp-vip.arn,
    //aws_alb_target_group.tgt-grp-ssh.arn,
  ]
  {{- end }}
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

  # todo volume tags
  # https://github.com/terraform-providers/terraform-provider-aws/issues/9448

  lifecycle {
    create_before_destroy = true
  }
}


resource "null_resource" "wait-{{ Dash ( Lower $v.Name ) }}" {
  depends_on = [
    "aws_autoscaling_group.{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}",
    "data.aws_instance.{{- Dash $v.Name }}" 
  ]

  count       = "{{ $v.Count }}"

  connection {
    timeout     = "{{ $v.ConnectionTimeout }}"
    user        = "{{ $.Username }}"
    private_key = "${var.private_key}"
    host        =
      {{- if $.ConfigureFromPrivateNet -}}
        element( data.aws_instance.{{- Dash $v.Name }}.*.private_ip, count.index )
      {{- else -}}
        element( data.aws_instance.{{- Dash $v.Name }}.*.public_ip, count.index )
      {{- end }}
  }

  provisioner "file" {
    content      = "terraform was able to ssh to the instance"
    destination = "/tmp/terraform.up"
  }
}

{{ end }}

# AWS Load Balancers
# ==============================================================================

resource "aws_alb" "alb" {
  name    = "{{ Dash ( Lower $.ClusterName ) }}"
  subnets = [{{ ( QuoteList ( AllSubNets ) ) }}]
  internal                   = false
  load_balancer_type         = "network"
  enable_deletion_protection = false

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = data.aws_caller_identity.current.arn
  }
}

resource "aws_alb_target_group" "tgt-grp-api" {
  depends_on = ["aws_alb.alb"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-api"
  port       = "{{ $.KubeAPISSLPort }}"
  protocol   = "TCP"
  vpc_id     = data.aws_vpc.vpc.id

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = data.aws_caller_identity.current.arn
  }
}

resource "aws_alb_target_group" "tgt-grp-vip" {
  depends_on = ["aws_alb.alb"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-vip"
  port       = "{{ $.KubeVIPAPISSLPort }}"
  protocol   = "TCP"
  vpc_id     = data.aws_vpc.vpc.id

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = data.aws_caller_identity.current.arn
  }
}

# resource "aws_alb_target_group" "tgt-grp-ssh" {
#   depends_on = ["aws_alb.alb"]
#   name       = "{{ Dash ( Lower $.ClusterName ) }}-ssh"
#   port       = 22
#   protocol   = "TCP"
#   vpc_id     = data.aws_vpc.vpc.id

#   tags = {
#     Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
#     Owner   = data.aws_caller_identity.current.arn
#   }
# }

resource "aws_alb_listener" "kube-api-ssl-port" {
  depends_on        = ["aws_alb.alb"]
  load_balancer_arn = "${aws_alb.alb.arn}"
  port              = "{{ $.KubeAPISSLPort }}"
  protocol          = "TCP"

  default_action {
    target_group_arn = aws_alb_target_group.tgt-grp-api.arn
    type             = "forward"
  }
}

resource "aws_alb_listener" "kube-vip-api-ssl-port" {
  depends_on        = ["aws_alb.alb"]
  load_balancer_arn = aws_alb.alb.arn
  port              = "{{ $.KubeVIPAPISSLPort }}"
  protocol          = "TCP"

  default_action {
    target_group_arn = aws_alb_target_group.tgt-grp-api.arn
    type             = "forward"
  }
}

# resource "aws_alb_listener" "ssh-listener" {
#   depends_on        = ["aws_alb.alb"]
#   load_balancer_arn = aws_alb.alb.arn
#   port              = 22
#   protocol          = "TCP"

#   default_action {
#     target_group_arn = aws_alb_target_group.tgt-grp-ssh.arn
#     type             = "forward"
#   }
# }

# AWS IAM
# ==============================================================================



resource "aws_key_pair" "keypair" {
  // TODO need to verify if key name change will cause destruction on existing 1.0 systems
  key_name   = "{{ Dash ( Lower $.ClusterName ) }}-key"
  public_key = "{{ Trim .PublicKey }}"
}

{{ range $k, $v := .NodePools }}

resource "aws_iam_instance_profile" "kube-{{ Dash ( Lower $v.Name ) }}-profile" {
  depends_on = ["aws_iam_role.kube-{{ Dash ( Lower $v.Name ) }}-role"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $v.Name ) }}-profile"
  role       = aws_iam_role.kube-{{ Dash ( Lower $v.Name ) }}-role.name
}

resource "aws_iam_role" "kube-{{ Dash ( Lower $v.Name ) }}-role" {
  name = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $v.Name ) }}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": { "Service": "ec2.amazonaws.com" },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "kube-{{ Dash ( Lower $v.Name ) }}-policy" {
  name       = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $v.Name ) }}-policy"
  role       = aws_iam_role.kube-{{ Dash ( Lower $v.Name ) }}-role.id
  depends_on = ["aws_iam_role.kube-{{ Dash ( Lower $v.Name ) }}-role"]

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {{ if eq $v.Name $masterNodePool.Name }}
    {
      "Effect": "Allow",
      "Action": [ "ec2:*" ],
      "Resource": [ "*" ]
    },
    {
      "Effect": "Allow",
      "Action": [ "elasticloadbalancing:*" ],
      "Resource": [ "*" ]
    },
    {{ else }}
    {
      "Effect": "Allow",
      "Action": "ec2:Describe*",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ec2:AttachVolume",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ec2:DetachVolume",
      "Resource": "*"
    },
    {{ end }}
    {
      "Effect": "Allow",
      "Action": [ "route53:*" ],
      "Resource": [ "*" ]
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [ "arn:aws:s3:::kubernetes-*" ]
    }
  ]
}
EOF
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