# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# resources.tf collects creates the resources that will be used with the image.  
# Be careful with what you create as a resource, as you can overwrite existing 
# infrastructure easily.

# AWS Instances
# ==============================================================================

{{ range $k, $v := .NodePools }}
  {{ if and $v.PGStrategy (isPGStrategy $v.PGStrategy) -}} 

resource "aws_placement_group" "node-pool-{{ Dash ( Lower $v.Name ) }}" {
  name     = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}"

  strategy = "{{ $v.PGStrategy }}"
}
  {{- end }}

resource "aws_instance" "{{ $v.Name }}" {
  depends_on = ["aws_iam_instance_profile.kube_{{ $k }}_profile"]

  connection {
    timeout     = "{{ $v.ConnectionTimeout }}"
    user        = "{{ $.Username }}"
    private_key = "${var.private_key}"
    host        = "
      {{- if $.ConfigureFromPrivateNet -}}
        ${self.private_ip} 
      {{- else -}}
        ${self.public_ip}
      {{- end }}"
  }
  {{ if $v.PGStrategy -}}
  placement_group        = "
    {{- if isPGStrategy $v.PGStrategy -}}
      {{- Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) -}}
    {{- else -}}
      {{- $v.PGStrategy -}}
    {{- end -}}"
  {{- end }}

  ami                    = "{{ $v.Ami }}"
  count                  = "{{ $v.Count }}"
  iam_instance_profile   = "${aws_iam_instance_profile.kube_{{ $k }}_profile.name}"
  instance_type          = "{{ $v.InstanceType }}"
  key_name               = "{{ Dash ( Lower $.ClusterName ) }}_key_{{ $.AwsEnv }}"
  subnet_id              = "${element(list({{ QuoteList $v.Subnets }}), count.index)}"
  vpc_security_group_ids = [{{ QuoteList $v.SecurityGroups }}]

  tags = {
    Name              = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}"
    NodePool          = "{{ Dash ( Lower $k ) }}"
    ClusterName       = "{{ Dash ( Lower $.ClusterName ) }}"
    "kubernetes.io/cluster/{{ Dash ( Lower $.ClusterName ) }}" = "owned"
  }

  volume_tags = {
    Name              = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}"
    NodePool          = "{{ Dash ( Lower $k ) }}"
    ClusterName       = "{{ Dash ( Lower $.ClusterName ) }}"
  }

  root_block_device {
    delete_on_termination = true
    volume_size           = "{{ $v.RootVolSize }}"
    volume_type           = "{{ $v.RootVolType }}"
  }

  provisioner "file" {
    content      = "terraform was able to ssh to the instance'"
    destination = "/tmp/terraform.up"
  }
}
{{ end }}

# AWS Load Balancers
# ==============================================================================

resource "aws_alb" "alb" {
  name    = "{{ Dash ( Lower $.ClusterName ) }}"
  subnets = [{{ ( QuoteList ( AllSubNets $.NodePools ) ) }}]
  internal                   = false
  load_balancer_type         = "network"
  enable_deletion_protection = false

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = "${data.aws_caller_identity.current.arn}"
  }
}

resource "aws_alb_target_group" "tgt-grp-api" {
  depends_on = ["aws_alb.alb"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-api"
  port       = "{{ $.KubeAPISSLPort }}"
  protocol   = "TCP"
  vpc_id     = "${data.aws_vpc.vpc.id}"

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = "${data.aws_caller_identity.current.arn}"
  }
}

resource "aws_alb_target_group" "tgt-grp-vip" {
  depends_on = ["aws_alb.alb"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-vip"
  port       = "{{ $.KubeVIPAPISSLPort }}"
  protocol   = "TCP"
  vpc_id     = "${data.aws_vpc.vpc.id}"

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = "${data.aws_caller_identity.current.arn}"
  }
}

resource "aws_alb_target_group" "tgt-grp-ssh" {
  depends_on = ["aws_alb.alb"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}-ssh"
  port       = 22
  protocol   = "TCP"
  vpc_id     = "${data.aws_vpc.vpc.id}"

  tags = {
    Cluster = "{{ Dash ( Lower $.ClusterName ) }}"
    Owner   = "${data.aws_caller_identity.current.arn}"
  }
}

resource "aws_alb_listener" "kube_api_ssl_port" {
  depends_on        = ["aws_alb.alb"]
  load_balancer_arn = "${aws_alb.alb.arn}"
  port              = "{{ $.KubeAPISSLPort }}"
  protocol          = "TCP"

  default_action {
    target_group_arn = "${aws_alb_target_group.tgt-grp-api.arn}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "kube_vip_api_ssl_port" {
  depends_on        = ["aws_alb.alb"]
  load_balancer_arn = "${aws_alb.alb.arn}"
  port              = "{{ $.KubeVIPAPISSLPort }}"
  protocol          = "TCP"

  default_action {
    target_group_arn = "${aws_alb_target_group.tgt-grp-api.arn}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "ssh-listener" {
  depends_on        = ["aws_alb.alb"]
  load_balancer_arn = "${aws_alb.alb.arn}"
  port              = 22
  protocol          = "TCP"

  default_action {
    target_group_arn = "${aws_alb_target_group.tgt-grp-ssh.arn}"
    type             = "forward"
  }
}

{{ $masterNodePool := MasterPool $.NodePools }}

resource "aws_alb_target_group_attachment" "app_tg_att_api" {
  depends_on       = ["aws_alb.alb"]
  count            = "{{ $masterNodePool.Count  }}"
  target_group_arn = "${aws_alb_target_group.tgt-grp-api.arn}"
  target_id        = "${element(aws_instance.{{ $masterNodePool.Name  }}.*.id, count.index)}"
  port             = "{{ $.KubeAPISSLPort }}"
}

resource "aws_alb_target_group_attachment" "app_tg_att_vip" {
  depends_on       = ["aws_alb.alb"]
  count            = "{{ $masterNodePool.Count  }}"
  target_group_arn = "${aws_alb_target_group.tgt-grp-vip.arn}"
  target_id        = "${element(aws_instance.{{ $masterNodePool.Name  }}.*.id, count.index)}"
  port             = "{{ $.KubeVIPAPISSLPort }}"
}

resource "aws_alb_target_group_attachment" "app_tg_att_ssh" {
  depends_on       = ["aws_alb.alb"]
  count            = "{{ $masterNodePool.Count  }}"
  target_group_arn = "${aws_alb_target_group.tgt-grp-ssh.arn}"
  target_id        = "${element(aws_instance.{{ $masterNodePool.Name  }}.*.id, count.index)}"
  port             = 22
}

# AWS IAM
# ==============================================================================



resource "aws_key_pair" "keypair" {
  // TODO need to verify if key name change will cause destruction on existing 1.0 systems
  key_name   = "{{ Dash ( Lower $.ClusterName ) }}_key_{{ $.AwsEnv }}"
  public_key = "{{ Trim .PublicKey }}"
}

{{ range $k, $v := .NodePools }}

resource "aws_iam_instance_profile" "kube_{{ Dash ( Lower $k ) }}_profile" {
  depends_on = ["aws_iam_role.kube_{{ Dash ( Lower $k ) }}_role"]
  name       = "{{ Dash ( Lower $.ClusterName ) }}_{{ Dash ( Lower $k ) }}_profile"
  role       = "${aws_iam_role.kube_{{ Dash ( Lower $k ) }}_role.name}"
}

resource "aws_iam_role" "kube_{{ Dash ( Lower $k ) }}_role" {
  name = "{{ Dash ( Lower $.ClusterName ) }}_{{ Dash ( Lower $k ) }}_role"

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

resource "aws_iam_role_policy" "kube_{{ Dash ( Lower $k ) }}_policy" {
  name       = "{{ Dash ( Lower $.ClusterName ) }}_{{ Dash ( Lower $k ) }}_policy"
  role       = "${aws_iam_role.kube_{{ Dash ( Lower $k ) }}_role.id}"
  depends_on = ["aws_iam_role.kube_{{ Dash ( Lower $k ) }}_role"]

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
