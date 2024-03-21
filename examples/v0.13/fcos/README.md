# Fedora CoreOS multi-machine example setup

### Requirements
* Linux Kernel ~> 4.14
* Libvirt ~> 3.0
* QEMU ~> 2.6

This is a relatively simple scalable example using Fedroa CoreOS as operating system.
By modifying the `hosts` variable you can start any number of virtual machines
with their own Ignition configuration

This example will work on x86\_64 (AMD64). The Terraform provider uses QEMU's firmware configuration (fw\_cfg) device to pass the Ignition config through to the Ignition instance running in the virtual machine. In order for this to work, Ignition needs to be run with `--provider=qemu`.


### Using the QEMU Guest Agent

In case you don't use the networks provided by libvirt you may run into the issue that you won't be able to receive the IP addresses from the VM you create.

Using the QEMU guest agent allows libvirt to pick up the address by hooking itself into the guest operating system.
As Fedora CoreOS comes without any guest agents we need to supply it from somewhere.
If the machine has internet access you can edit the `qemu-agent.service` file and remove the `ExecStartPre` line and the docker daemon should download the appropriate container when you activate the service file in the Ignition config. If the machine has no access to the internet we need to upload the container from the KVM host.[1]
```bash
$ docker pull docker.io/rancher/os-qemuguestagent:v2.8.1-2
$ docker save docker.io/rancher/os-qemuguestagent:v2.8.1-2 -o /srv/images/qemu-guest-agent.tar
```

Make sure the relevant blocks are uncommented in the domain definition and the Ignition config. The Ignition configuration should include the two additional files `docker-images.mount` and `qemu-agent.service`. Note that the`qemu-guest-agent.tar` needs to be local to the KVM host and not the machine running terraform. Also the /srv/images directory needs to be readable by whatever user your qemu process runs as.


### Known Bugs
1. Before Linux 4.14-rc2 the graphics option "autoport" will not work and libvirt will try to create all machines with the same Spice/VNC port
2. Below libvirt v3 the generated Ignition id will change when the number of machines is changed causing a destroy/create for all machines.


[1]: Based on the work of [@tommyknows](https://github.com/dmacvicar/terraform-provider-libvirt/issues/364#issuecomment-442164364) and [@remoe](https://github.com/dmacvicar/terraform-provider-libvirt/issues/364#issuecomment-443456552).
