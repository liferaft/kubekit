# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# resources.tf collects creates the resources that will be used with the image.  
# Be careful with what you create as a resource, as you can overwrite existing 
# infrastructure easily.
{{ range $k, $v := .NodePools }}

# for backward compat, master and worker need to be dumb-master, dumb-worker ?
# TODO TEST UPDATE ON OLD CLUSTER
resource "vsphere_virtual_machine" "{{ Dash ( Lower $v.Name ) }}" {
  count = "{{ $v.Count }}"

  name = lookup({{ ExtractAddressPoolToTFIndexMap $v.AddressPool "hostname" }}, count.index, "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}")

  resource_pool_id = data.vsphere_resource_pool.pool.id
  datastore_id     = data.vsphere_datastore.datastore.id

  folder   = "{{ $.Folder }}"
  num_cpus = "{{ $v.CPUs }}"
  memory   = "{{ $v.Memory }}"
  guest_id = data.vsphere_virtual_machine.{{- Dash ( Lower $v.Name ) -}}-template.guest_id

  scsi_type        = data.vsphere_virtual_machine.{{- Dash ( Lower $v.Name ) -}}-template.scsi_type
  enable_disk_uuid = "true"

  network_interface {
    network_id   = data.vsphere_network.network.id
    adapter_type = data.vsphere_virtual_machine.{{- Dash ( Lower $v.Name ) -}}-template.network_interface_types[0]
  }

  # leaving at single disk now, but templating will allow for multiples
  disk {
    label            = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}.vmdk"
    size             = "{{ $v.RootVolSize }}"
    eagerly_scrub    = data.vsphere_virtual_machine.{{ Dash ( Lower $v.Name ) }}-template.disks.0.eagerly_scrub
    thin_provisioned = data.vsphere_virtual_machine.{{ Dash ( Lower $v.Name ) }}-template.disks.0.thin_provisioned
    unit_number      = 0
  }

  extra_config = {
    "guestinfo.cloudinit.userdata" = "#cloud-config\nhostname: {{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}\n\nssh_authorized_keys:\n  - \"{{ Trim $.PublicKey }}\n\""
  }

  clone {
    linked_clone  = "{{ $v.LinkedClone }}"
    template_uuid = data.vsphere_virtual_machine.{{ Dash ( Lower $v.Name ) }}-template.id

    customize {
      network_interface {
        ipv4_address = lookup({{ ExtractAddressPoolToTFIndexMap $v.AddressPool "ip" }}, count.index, "")

        {{ if $v.IPNetmask }}
        ipv4_netmask = "{{ $v.IPNetmask }}"
        {{ end }}
      }

      linux_options {
        host_name = lookup({{ ExtractAddressPoolToTFIndexMap $v.AddressPool "hostname" }}, count.index, "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}")
        domain    = "{{ $.Domain }}"
      }

      {{ if ne $v.IPGateway "" }}
      ipv4_gateway    = "{{ $v.IPGateway }}"
      {{ end }}

      dns_server_list = [{{ QuoteList $.DNSServers }}]
    }
  }
}

{{ end }}
