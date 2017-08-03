# Example Configs

Each of these examples is written in version 2.1.0 of the config. Ensure that any configuration is compatible with the version that Ignition accepts. Compatibility requires the major versions to match and the spec version be less than or equal to the version Ignition accepts.

## Services

### Start Services

This config will write a single service unit (shown below) with the contents of an example service. This unit will be enabled as a dependency of multi-user.target and therefore start on boot.

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "systemd": {
    "units": [{
      "name": "example.service",
      "enabled": true,
      "contents": "[Service]\nType=oneshot\nExecStart=/usr/bin/echo Hello World\n\n[Install]\nWantedBy=multi-user.target"
    }]
  }
}
```

#### example.service

```INI
[Service]
Type=oneshot
ExecStart=/usr/bin/echo Hello World

[Install]
WantedBy=multi-user.target
```

### Modify Services

This config will add a [systemd unit drop-in](https://coreos.com/os/docs/latest/using-systemd-drop-in-units.html) to modify the existing service `systemd-networkd` and sets its environment variable `SYSTEMD_LOG_LEVEL` to `debug`.

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "systemd": {
    "units": [{
      "name": "systemd-networkd.service",
      "dropins": [{
        "name": "debug.conf",
        "contents": "[Service]\nEnvironment=SYSTEMD_LOG_LEVEL=debug"
      }]
    }]
  }
}
```

#### systemd-networkd.service.d/debug.conf

```INI
[Service]
Environment=SYSTEMD_LOG_LEVEL=debug
```

## Reformat the Root Filesystem

This example Ignition configuration will locate the device with the "ROOT" filesystem label (the root filesystem) and reformat it to btrfs, recreating the filesystem label. The `force` option is set to ensure that `mkfs.btrfs` ignores any existing filesystem.

### Btrfs

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "filesystems": [{
      "mount": {
        "device": "/dev/disk/by-label/ROOT",
        "format": "btrfs",
        "wipeFilesystem": true,
        "options": [ "--label=ROOT" ]
      }
    }]
  }
}
```

### XFS

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "filesystems": [{
      "mount": {
        "device": "/dev/disk/by-label/ROOT",
        "format": "xfs",
        "wipeFilesystem": true,
        "options": [ "-L", "ROOT" ]
      }
    }]
  }
}
```

The create options are forwarded to the underlying `mkfs.$format` utility. The respective `mkfs.$format` manual pages document the available options.

## Create Files on the Root Filesystem

In many cases it is useful to write files to the root filesystem. This example writes a single file to `/foo/bar` on the root filesystem. The contents of the file ("example file") are specified inline in the config using the [data URL scheme][rfc2397].

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/foo/bar",
      "contents": { "source": "data:,example%20file%0A" }
    }]
  }
}
```

The config makes use of the universally-defined "root" filesystem. This filesystem is defined within Ignition itself and roughly looks like the following. The "root" filesystem allows additional configs to reference the root filesystem, regardless of its type (e.g. btrfs, tmpfs, ext4).

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "filesystems": [{
      "name": "root",
      "path": "/sysroot"
    }]
  }
}
```

## Create Files from Remote Contents

There are cases where it is desirable to write a file to disk, but with the contents of a remote resource. The following config demonstrates how to do this in addition to validating the contents of the file.

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/foo/bar",
      "contents": {
        "source": "http://example.com/asset",
        "verification": { "hash": "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" }
      }
    }]
  }
}
```

The SHA512 sum of the file can be determined using `sha512sum`.

## Create a RAID-enabled Data Volume

In many scenarios, it may be useful to have an external data volume. This config will set up a RAID0 ext4 volume, `data`, between two separate disks. It also writes a mount unit (shown below) which will automatically mount the volume to `/var/lib/data`.

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "disks": [
      {
        "device": "/dev/sdb",
        "wipeTable": true,
        "partitions": [{
          "label": "raid.1.1",
          "number": 1,
          "size": 20480,
          "start": 0
        }]
      },
      {
        "device": "/dev/sdc",
        "wipeTable": true,
        "partitions": [{
          "label": "raid.1.2",
          "number": 1,
          "size": 20480,
          "start": 0
        }]
      }
    ],
    "raid": [{
      "devices": [
        "/dev/disk/by-partlabel/raid.1.1",
        "/dev/disk/by-partlabel/raid.1.2"
      ],
      "level": "stripe",
      "name": "data"
    }],
    "filesystems": [{
      "mount": {
        "device": "/dev/md/data",
        "format": "ext4",
        "create": { "options": [ "-L", "DATA" ] }
      }
    }]
  },
  "systemd": {
    "units": [{
      "name": "var-lib-data.mount",
      "enable": true,
      "contents": "[Mount]\nWhat=/dev/md/data\nWhere=/var/lib/data\nType=ext4\n\n[Install]\nWantedBy=local-fs.target"
    }]
  }
}
```

### var-lib-data.mount

```INI
[Mount]
What=/dev/md/data
Where=/var/lib/data
Type=ext4

[Install]
WantedBy=local-fs.target
```

## Replace the Config with a Remote Config

In some cloud environments, there is a limit on the size of the config which may be provided to a machine. To work around this, Ignition allows configs to be replaced with the contents of an alternate, remote config. The following demonstrates this, using a SHA512 sum to verify the contents of the config.

```json ignition
{
  "ignition": {
    "version": "2.1.0",
    "config": {
      "replace": {
        "source": "http://example.com/config.json",
        "verification": { "hash": "sha512-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" }
      }
    }
  }
}
```

The SHA512 sum of the config can be determined using `sha512sum`.

## Set the Hostname

Setting the hostname of a system is as simple as writing `/etc/hostname`:

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/etc/hostname",
      "mode": 420,
      "contents": { "source": "data:,core1" }
    }]
  }
}
```

## Add Users

Users can be added to an OS with the `passwd.users` key which takes a list of objects that specify a given user. If you wanted to configure a user "systemUser" and a user "jenkins" you would do that as follows:

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "passwd": {
    "users": [
      {
        "name": "systemUser",
        "passwordHash": "$superSecretPasswordHash.",
        "sshAuthorizedKeys": [
          "ssh-rsa veryLongRSAPublicKey"
        ]
      },
      {
        "name": "jenkins",
        "uid": 1000
      }
    ]
  }
}
```

To add more users, configure them within the `users` list structure (`[...]`).

[rfc2397]: http://tools.ietf.org/html/rfc2397
