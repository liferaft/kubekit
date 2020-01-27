# Variables
# ==============================================================================

variable "access_key" {}
variable "secret_key" {}
variable "region"   {}
variable "token"   {}

locals {

{{ range $k, $v := .NodePools }}
  node-{{ Dash ( Lower $v.Name ) }}-userdata =  <<USERDATA
#!/bin/bash
ulimit -n 65535
cat <<EOF > /etc/security/limits.d/30-nofile.conf
root       soft    nofile     65536
root       hard    nofile     65536
EOF
cat <<EOF > /etc/sysctl.d/99-kubelet-network.conf
# Have a larger connection range available
net.ipv4.ip_local_port_range=1024 65000

# Reuse closed sockets faster
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_fin_timeout=15

# The maximum number of "backlogged sockets".  Default is 128.
net.core.somaxconn=4096
net.core.netdev_max_backlog=4096

# 16MB per socket - which sounds like a lot,
# but will virtually never consume that much.
net.core.rmem_max=16777216
net.core.wmem_max=16777216

# Various network tunables
net.ipv4.tcp_max_syn_backlog=20480
net.ipv4.tcp_max_tw_buckets=400000
net.ipv4.tcp_no_metrics_save=1
net.ipv4.tcp_rmem=4096 87380 16777216
net.ipv4.tcp_syn_retries=2
net.ipv4.tcp_synack_retries=2
net.ipv4.tcp_wmem=4096 65536 16777216
#vm.min_free_kbytes=65536

# Connection tracking to prevent dropped connections (usually issue on LBs)
net.netfilter.nf_conntrack_max=262144
net.ipv4.netfilter.ip_conntrack_generic_timeout=120
net.netfilter.nf_conntrack_tcp_timeout_established=86400

# ARP cache settings for a highly loaded docker swarm
net.ipv4.neigh.default.gc_thresh1=8096
net.ipv4.neigh.default.gc_thresh2=12288
net.ipv4.neigh.default.gc_thresh3=16384

# disable ipv6
net.ipv6.conf.all.disable_ipv6=1
net.ipv6.conf.default.disable_ipv6=1
net.ipv6.conf.lo.disable_ipv6=1

# set max_map_count for elasticsearch usage
vm.max_map_count={{ $.MaxMapCount }}

fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=524288
EOF
systemctl restart systemd-sysctl.service
sed -i  's/dateext.*/dateext dateformat -%Y-%m-%d-%s.log/' /etc/logrotate.conf
set -o xtrace
systemctl stop kubelet
  {{- if IsFastEphemeral $v }}
yum install rsync -y
pvcreate /dev/nvme[1-9]*n*
vgcreate vgdata /dev/nvme[1-9]*n*
lvcreate -l 100%FREE --type striped -n lvoldata vgdata
mkfs.xfs /dev/vgdata/lvoldata
mkdir -p /data
mount /dev/vgdata/lvoldata /data
echo "/dev/vgdata/lvoldata /data xfs defaults 0 0"  >> /etc/fstab
for directory in /var/lib/docker /var/lib/kubelet; do
  if [ ! -d "$${directory}" ]; then
    mkdir -p $${directory}
  fi
  if [ ! -z "$(ls -A $${directory})" ]; then
    rsync -avzh $${directory}/ /data/$(basename $${directory})
  fi
  rm -Rf $${directory}/*
  mount --bind /data/$(basename $${directory}) $${directory}
  echo "$${directory} /data/$(basename $${directory}) bind bind 0 0"  >> /etc/fstab
done
  {{- end }}
/etc/eks/bootstrap.sh --kubelet-extra-args '--node-labels="
{{- if ne ( len $v.KubeletNodeLabels ) 0 -}}{{- Join $v.KubeletNodeLabels "," -}}
{{- else -}}{{- Join $.DefaultNodePool.KubeletNodeLabels "," -}}
{{- end -}}" --register-with-taints="
{{- if ne ( len $v.KubeletNodeTaints ) 0 -}}{{- Join $v.KubeletNodeTaints "," -}}
{{- else -}}{{- Join $.DefaultNodePool.KubeletNodeTaints "," -}}
{{- end -}}"' --apiserver-endpoint '${aws_eks_cluster.kubekit.endpoint}' --b64-cluster-ca '${aws_eks_cluster.kubekit.certificate_authority.0.data}' '{{ $.ClusterName }}'
USERDATA

{{ end }}
}
