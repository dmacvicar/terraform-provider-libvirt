## Flatcar Linux simple cluster setup

### Requirements

This setup works with the following config:
```shell
$ libvirtd --version
libvirtd (libvirt) 5.5.0
$ terraform version
Terraform v0.12.6
```

### Flatcar Linux

Ths example is strongly inspired by the CoreOS [example](https://github.com/dmacvicar/terraform-provider-libvirt/tree/master/examples/v0.12/coreos).

QEMU-agent is not used, network is configured with static IP to keep things simple.

```shell
$ virsh net-dumpxml --network cluster-net
...
  <ip address='192.168.122.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.122.100' end='192.168.122.254'/>
      <host mac='52:54:00:00:00:a1' name='node-01' ip='192.168.122.101'/>
      <host mac='52:54:00:00:00:a2' name='node-02' ip='192.168.122.102'/>
      <host mac='52:54:00:00:00:a3' name='node-03' ip='192.168.122.103'/>
    </dhcp>
  </ip>
...
```

Do not forget to download Flatcar Linux image and to add it the pool of your [choice](https://docs.flatcar-linux.org/os/booting-with-libvirt/#choosing-a-channel).

You can specify the number of hosts by updating: 
```hcl
variable "hosts" {
  default = 2
}
```

(:warning: you will need to update `units/etcd-member.conf` to add or remove or node to the initial cluster :warning:)

Add you SSH pub key or provide a hashed password.

Finally, `fw_cfg_name` is the "path" where ignition file will be mounted [doc](https://docs.flatcar-linux.org/os/booting-with-libvirt/#creating-the-domain-xml).

```hcl
resource "libvirt_domain" "node" {
  ...
  fw_cfg_name = "opt/org.flatcar-linux/config"
  ...
}
```

