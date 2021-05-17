# Output the private and public IPs of the instance

output "LinuxVMPrivateIPs" {
  value = [oci_core_instance.LinuxVM.*.private_ip]
}

output "LinuxVMPublicIP" {
  value = [oci_core_instance.LinuxVM.*.public_ip]
}
