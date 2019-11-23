{{ $masterNodePool := MasterPool $.NodePools }}

output "service_ip" {
  value = "
{{- if $.KubeVirtualIPApi -}}
  {{- $.KubeVirtualIPApi -}} 
{{- else -}}  
  ${openstack_compute_floatingip_associate_v2.float_assoc-{{ Dash ( Lower $masterNodePool.Name ) }}.0.floating_ip}
{{- end }}"
}

output "service_port" {
  value = "
{{- if and $.KubeVirtualIPApi $.KubeVIPAPISSLPort -}}
  {{- $.KubeVIPAPISSLPort -}} 
{{- else -}} 
  {{- $.KubeAPISSLPort -}}
{{- end }}"
}

output "nodes" {
 	value = [ {{- range $k, $v := $.NodePools -}} {{- range $i := Count $v.Count  }}
    "{\"private_ip\": \"${openstack_compute_instance_v2.
    {{- Dash ( Lower $v.Name ) }}.{{ $i }}.access_ip_v4}\",\"public_ip\": \"${openstack_compute_floatingip_associate_v2.float_assoc-
    {{- Dash ( Lower $k ) }}.{{ $i }}.floating_ip}\",\"public_dns\": \"${openstack_compute_instance_v2.
    {{- Dash ( Lower $v.Name ) }}.{{ $i }}.name}\",\"private_dns\": \"${openstack_compute_instance_v2.
    {{- Dash ( Lower $v.Name ) }}.{{ $i }}.name}\",\"pool\": \"{{ $v.Name }}\",\"role\": \"{{ Dash ( Lower $k ) }}\"}",{{ end }}{{ end }}
  ]
}