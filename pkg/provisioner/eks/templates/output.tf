# Outputs
# ==============================================================================
output "endpoint" {
  value = aws_eks_cluster.kubekit.endpoint
}

output "certificate-authority" {
  value = aws_eks_cluster.kubekit.certificate_authority.0.data
}

output "role-arn" {
  value = aws_iam_role.cluster-node.arn
}

output "kubernetes_version" {
  value = aws_eks_cluster.kubekit.version
}

{{ range $k, $v := $.NodePools }}
output "{{ $v.Name }}-ami" {
  value = aws_launch_configuration.node-{{ Dash ( Lower $v.Name ) }}.image_id
}
{{ end }}

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
}