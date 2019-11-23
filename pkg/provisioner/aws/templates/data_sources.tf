# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# data_sources.tf collects data and set's variables to be used later.  
# It does nothing to modify the images

data "aws_vpc" "vpc" {
  id            = "{{ $.AwsVpcID }}"
}

data "aws_subnet_ids" "vpc_subnet" {
  vpc_id        = "{{ $.AwsVpcID }}"
}

data "aws_caller_identity" "current" {}

{{ range $k, $v := .NodePools }}

  {{ if gt $v.Count 0 }}
data "aws_instance" "{{ Dash ( Lower $k ) }}" {
  count = "{{ $v.Count }}"
  depends_on = ["data.aws_instances.{{ Dash ( Lower $k ) }}"]
  instance_id = "${data.aws_instances.{{ Dash ( Lower $k ) }}.ids[count.index]}"
}
  

data "aws_instances" "{{ Dash ( Lower $k ) }}" {
  depends_on = [ "aws_instance.{{ Dash ( Lower $v.Name ) }}",
  ]
  instance_tags = {
    NodePool          = "{{ Dash ( Lower $k ) }}"
    ClusterName       = "{{ Dash ( Lower $.ClusterName ) }}"
  }
}
  {{ end }}
{{ end }}
