{{ $masterNodePool := MasterPool $.NodePools }}

output "service_ip" {
  value = "${aws_instance.{{ $masterNodePool.Name  }}.0.public_ip}"
}

output "service_port" {
  value = "{{ $.KubeAPISSLPort }}"
}

output "alb_dns" {
  value = "${aws_alb.alb.dns_name}"
}

output "kube_vip_api_ssl_port" {
  value = "${aws_alb_listener.kube_vip_api_ssl_port.port}"
}

output "kube_api_ssl_port" {
  value = "${aws_alb_listener.kube_api_ssl_port.port}"
}

output "nodes" {
  value = [ {{- range $k, $v := $.NodePools -}} {{- range $i := Count $v.Count  }}
      "{\"private_ip\": \"${data.aws_instance.
      {{- Dash ( Lower $k ) }}.{{ $i }}.private_ip}\",\"public_ip\": \"${data.aws_instance.
      {{- Dash ( Lower $k ) }}.{{ $i }}.public_ip}\",\"public_dns\": \"${data.aws_instance.
      {{- Dash ( Lower $k ) }}.{{ $i }}.public_dns}\",\"private_dns\": \"${data.aws_instance.
      {{- Dash ( Lower $k ) }}.{{ $i }}.private_dns}\",\"pool\": \"{{ $v.Name }}\",\"role\": \"{{ Dash ( Lower $k ) }}\"}",{{ end }}{{ end }}
  ]
}
