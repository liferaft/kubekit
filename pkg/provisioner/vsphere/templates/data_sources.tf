# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

# data_sources.tf collects data and set's varaibles to be used later.  
# It does nothing to modify the images

data "vsphere_datacenter" "dc" {
  name = "{{ .Datacenter }}"
}

data "vsphere_datastore" "datastore" {
  name          = "{{ .Datastore }}"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_resource_pool" "pool" {
  name          = "{{ .ResourcePool }}"
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "network" {
  name          = "{{ .VsphereNet }}"
  datacenter_id = data.vsphere_datacenter.dc.id
}

{{ range $k, $v := .NodePools }}
data "vsphere_virtual_machine" "{{ Dash ( Lower $v.Name ) }}-template" {
  name          = "{{ $v.TemplateName }}"
  datacenter_id = data.vsphere_datacenter.dc.id
}
{{ end }}