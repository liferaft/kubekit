// resources.tf collects creates the resources that will be used with the image.  Be careful with what you create as a
// resource, as you can overwrite existing infrastructure easily.

resource "openstack_compute_keypair_v2" "keypair" {
  region     = "{{ $.OpenstackRegion }}"
  name       = "{{ Dash ( Lower $.ClusterName ) }}-keypair"
  public_key = "{{ Trim $.PublicKey }}\n"
}

{{ range $k, $v := .NodePools }}

resource "openstack_compute_instance_v2" "{{ Dash ( Lower $v.Name ) }}" {
  depends_on      = ["openstack_networking_floatingip_v2.float-{{ Dash ( Lower $k ) }}"]
  count           = "{{ $v.Count }}"
  name            = "{{ Dash ( Lower $.ClusterName ) }}-{{ Dash ( Lower $k ) }}-${format("%02d", count.index+1)}"
  image_id        = "{{ $v.OpenstackImageID }}"
  flavor_id       = "{{ $v.OpenstackFlavorID }}"
  key_pair        = "${openstack_compute_keypair_v2.keypair.id}"
  security_groups = [{{ QuoteList $v.SecurityGroups }}]

  // TODO: with templating, this can be extended create multiple networks and interfaces
  // Allowing for a hw / bynet like environment
  network {
    name           = "{{ $.OpenstackNetName }}"
    access_network = true
  }
}

resource "openstack_networking_floatingip_v2" "float-{{ Dash ( Lower $k ) }}" {
  count = "{{ $v.Count }}"
  pool  = "public"
}

resource "openstack_compute_floatingip_associate_v2" "float_assoc-{{ Dash ( Lower $k ) }}" {
  depends_on = ["openstack_compute_instance_v2.{{ Dash ( Lower $v.Name ) }}",
    "openstack_networking_floatingip_v2.float-{{ Dash ( Lower $k ) }}",
  ]

  count       = "{{ $v.Count }}"
  floating_ip = "${element(openstack_networking_floatingip_v2.float-{{ Dash ( Lower $k ) }}.*.address, count.index)}"
  instance_id = "${element(openstack_compute_instance_v2.{{ Dash ( Lower $v.Name ) }}.*.id, count.index)}"
}

resource "null_resource" "wait-{{ Dash ( Lower $k ) }}" {
  depends_on = [
    "openstack_compute_floatingip_associate_v2.float_assoc-{{ Dash ( Lower $k ) }}",
    "openstack_compute_instance_v2.{{ Dash ( Lower $v.Name ) }}",
  ]

  count       = "{{ $v.Count }}"

  connection {
    user        = "{{ $.Username }}"
    host        = "${element(openstack_networking_floatingip_v2.float-{{ Dash ( Lower $k ) }}.*.address, count.index)}"
    private_key = "${var.private_key}"
    timeout     = "5m"
  }

  provisioner "file" {
    content      = "terraform was able to ssh to the instance'"
    destination = "/tmp/terraform.up"
  }
}
{{ end }}
