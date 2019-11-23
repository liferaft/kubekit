# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

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
