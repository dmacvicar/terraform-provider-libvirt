# TODO

## Update acceptance tests for new disk schema

The disk schema was changed to follow libvirt XML API structure. Tests need to be updated from:

```hcl
disk = {
  volume_id = libvirt_volume.vm1_disk.id
}
```

To:

```hcl
disk = {
  source = {
    pool   = libvirt_volume.vm1_disk.pool
    volume = libvirt_volume.vm1_disk.name
  }
}
```

### Failing tests
- TestAccDomainResource_network
- TestAccDomainResource_directNetwork* (all variants)
- TestAccDomainResource_graphics
- TestAccDomainResource_video
- TestAccDomainResource_emulator
- TestAccDomainResource_console
- TestAccDomainResource_serial
- TestAccDomainResource_rng
- TestAccDomainResource_filesystem
- TestAccDomainResource_autostart
- TestAccDomainResource_diskWWN
- TestAccDomainResource_diskBlockDevice
- TestAccVolumeResource_withDomain

### Files to update
- internal/provider/domain_resource_test.go
- internal/provider/volume_resource_test.go
