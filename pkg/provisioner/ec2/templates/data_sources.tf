# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# data_sources.tf collects data and set's variables to be used later.  
# It does nothing to modify the images

data "aws_region" "current" {}

data "aws_vpc" "vpc" {
  id            = "{{ $.AwsVpcID }}"
}

data "aws_subnet_ids" "vpc_subnet" {
  vpc_id        = "{{ $.AwsVpcID }}"
}

data "aws_caller_identity" "current" {}

{{ range $k, $v := .NodePools }}

  {{ if gt $v.Count 0 }}
data "aws_instance" "{{ Dash ( Lower $v.Name ) }}" {
  count = "{{ $v.Count }}"
  depends_on = ["data.aws_instances.{{ Dash ( Lower $v.Name ) }}"]
  instance_id = data.aws_instances.{{ Dash ( Lower $v.Name ) }}.ids[count.index]
}
  

data "aws_instances" "{{ Dash ( Lower $v.Name ) }}" {
  depends_on = [ "aws_autoscaling_group.{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}",
  ]
  instance_tags = {
    Name = "{{ Dash ( Lower $.ClusterName ) }}-node-{{ Dash ( Lower $v.Name ) }}" 
  }
}
  {{ end }}
{{ end }}