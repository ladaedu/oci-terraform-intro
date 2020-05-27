# Output the private and public IPs of the instance

output "WebServerPrivateIPs" {
  value = [oci_core_instance.WebServer.*.private_ip]
}

#output "WebServerPublicIPs" {
#  value = ["${oci_core_instance.WebServer.*.public_ip}"]
#}

output "WebServerHostNames" {
  value = [oci_core_instance.WebServer.*.hostname_label]
}

output "WebServerDomain" {
  value = [oci_core_subnet.PrivateSubnet.*.subnet_domain_name]
}

output "BastionPublicIP" {
  value = [oci_core_instance.Bastion.*.public_ip]
}

output "VcnID" {
  value = [oci_core_virtual_network.VCN.id]
}
