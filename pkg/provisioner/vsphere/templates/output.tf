{{ $masterNodePool := MasterPool $.NodePools }}

output "service_ip" {
  value =
{{- if $.KubeVirtualIPApi -}}
  "{{- $.KubeVirtualIPApi -}}"
{{- else -}}  
  vsphere_virtual_machine.{{ Dash ( Lower $masterNodePool.Name ) }}.0.default_ip_address
{{- end }}
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
      "{\"private_ip\": \"${vsphere_virtual_machine.
      {{- Dash ( Lower $v.Name ) }}.{{ $i }}.default_ip_address}\",\"public_ip\": \"${vsphere_virtual_machine.
      {{- Dash ( Lower $v.Name ) }}.{{ $i }}.default_ip_address}\",\"public_dns\": \"${vsphere_virtual_machine.
      {{- Dash ( Lower $v.Name ) }}.{{ $i }}.name}\",\"private_dns\": \"${vsphere_virtual_machine.
      {{- Dash ( Lower $v.Name ) }}.{{ $i }}.name}\",\"pool\": \"{{ $v.Name }}\",\"role\": \"{{ Dash ( Lower $k ) }}\"}",{{ end }}{{ end }}
  ]
}