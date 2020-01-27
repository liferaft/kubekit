{{ $masterNodePool := MasterPool $.NodePools }}

output "service_ip" {
  value = aws_alb.alb.dns_name
}

output "service_port" {
  value = aws_alb_listener.kube-vip-api-ssl-port.port
}

output "alb_dns" {
  value = aws_alb.alb.dns_name
}

output "kube_vip_api_ssl_port" {
  value = aws_alb_listener.kube-vip-api-ssl-port.port
}

output "kube_api_ssl_port" {
  value = aws_alb_listener.kube-api-ssl-port.port
}

output "nodes" {
  value = [ {{- range $k, $v := $.NodePools -}} {{- range $i := Count $v.Count  }}
      "{\"private_ip\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.private_ip}\",\"public_ip\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.public_ip}\",\"public_dns\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.public_dns}\",\"private_dns\": \"${data.aws_instance.
      {{- Dash $v.Name }}.{{ $i }}.private_dns}\",\"role\": \"{{ Dash ( Lower $v.Name ) }}\",\"pool\": \"{{ Dash ( Lower $v.Name ) }}\"}",{{ end }}{{ end }}
  ]
}

output "elastic-fileshares" {
  value = "[ {{- $first := true -}}{{- range $k, $v := $.ElasticFileshares -}}
    {{- if $first -}}{{- $first = false -}}{{- else -}}, {{end -}}
    { \"efs_name\": \"{{- Dash ( Lower $k ) }}\",\"efs_id\": \"${aws_efs_file_system.efs-
    {{- Dash ( Lower $k ) }}.id}\",\"efs_dns\": \"${aws_efs_file_system.efs-
    {{- Dash ( Lower $k ) }}.dns_name}\",\"efs_region\": \"${data.aws_region.current.name}\" }
    {{- end }} ]"
}