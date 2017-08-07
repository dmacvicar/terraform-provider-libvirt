# Migrating Between Configuration Versions

Occasionally, there are changes made to Ignition's configuration that break backward compatibility. While this is not a concern for running machines (since Ignition only runs one time during first boot), it is a concern for those who maintain configuration files. This document serves to detail each of the breaking changes and tries to provide some reasoning for the change. This does not cover all of the changes to the spec - just those that need to be considered when migrating from one version to the next.

## From Version 2.0.0 to 2.1.0

There are not any breaking changes between versions 2.0.0 and versions 2.1.0 of the configuration specification. Any valid 2.0.0 configuration can be updated to a 2.1.0 configuration by simply changing the version string in the config.

The 2.1.0 version of the configuration is greatly improved over version 2.0.0, with many new fields and behaviors added to the specification.

The following is a list of notable new features, deprecations, and changes.

### HTTP timeouts

The values used to control the backoff when retrying in HTTP requests can now be set in the config. For details on how the backoff logic works, please see the section in the [operator's notes][operator-notes].

The fields to do this are in a new object called `timeouts`, and they can alter the time spent waiting for HTTP response headers and the total time limit for the operation.

```json ignition
{
  "ignition": {
    "version": "2.1.0",
    "timeouts": {
      "httpResponseHeaders": 20,
      "httpTotal": 600
    }
  }
}
```

### Partition GUIDs

The GPT unique partition GUID can now be set on partitions in a configuration.

```json ignition
{
  "ignition": {
    "version": "2.1.0"
  },
  "storage": {
    "disks": [
      {
        "device": "/dev/disk/by-uuid/ecdbeb92-174e-4d24-9d6f-fbd9cb668a48",
        "partitions": [
          {
            "guid": "8a7a6e26-5e8f-4cca-a654-46215d4696ac"
          }
        ]
      }
    ]
  }
}
```

### Directories, links, and files

Version 2.1.0 of the configuration specification now supports specifying directories and links (both symbolic and hard) to be created, and when creating either of these or creating a file the owner's user and group can be specified by name in addition to UID and GID.

```json ignition
{
  "ignition": { "version": "2.1.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/home/core/foo.txt",
      "mode": 420,
      "contents": { "source": "data:,helloworld" },
      "user": {
        "name": "core"
      },
      "group": {
        "name": "core"
      }
    }],
    "directories": [{
      "filesystem": "root",
      "path": "/home/core/bar",
      "mode": 493,
      "user": {
        "name": "core"
      },
      "group": {
        "name": "core"
      }
    }],
    "links": [{
      "filesystem": "root",
      "path": "/home/core/baz.txt",
      "target": "/home/core/foo.txt",
      "hard": true,
      "user": {
        "name": "core"
      },
      "group": {
        "name": "core"
      }
    }]
  }
}
```

### Filesystem create object deprecation

Version 2.0.0 of the configuration specification included an object named `create` in the `mount` object under the `filesystems` section.

```json ignition
{
  "ignition": {
    "version": "2.1.0"
  },
  "storage": {
    "filesystems": [
      {
        "name": "data",
        "mount": {
          "device": "/dev/disk/by-uuid/ecdbeb92-174e-4d24-9d6f-fbd9cb668a48",
          "format": "ext4",
          "create": {
            "force": true,
            "options": ["-L", "DATA", "-b", "1024"]
          }
        }
      }
    ]
  }
}
```

This `create` object has been deprecated. Configurations of version 2.1.0 that use the `create` object will still work, but will cause Ignition to log a warning.

It is now advised to use the new fields that have been added to the `mount` object.

```json ignition
{
  "ignition": {
    "version": "2.1.0"
  },
  "storage": {
    "filesystems": [
      {
        "name": "data",
        "mount": {
          "device": "/dev/disk/by-uuid/ecdbeb92-174e-4d24-9d6f-fbd9cb668a48",
          "format": "ext4",
          "wipeFilesystem": true,
          "label": "DATA",
          "options": ["-b", "1024"]
        }
      }
    ]
  }
}
```

The `wipeFilesystem` flag that replaces the `force` flag has rather different semantics, and can allow for existing filesystems to be reused. For more information, please see the [filesystems document][filesystems].

### Passwd create object deprecation

Similar to the `create` object in the `filesystems` section, version 2.0.0 of the configuration specification included an object named `create` in the `users` list in the `passwd` object.

```json ignition
{
  "ignition": {
    "version": "2.1.0"
  },
  "passwd": {
    "users": [
      {
        "name": "test",
        "create": {
          "uid": 1010,
          "gecos": "user creation test",
          "noCreateHome": true,
          "noUserGroup": true
        },
      }
    ]
  }
}
```

This `create` object has been deprecated. Configurations of version 2.1.0 that use the `create` object will still work, but will cause Ignition to log a warning.

The fields that existed in the `create` object have been added directly to the `users` object, and it's advised to use these new fields instead of the `create` object.

```json ignition
{
  "ignition": {
    "version": "2.1.0"
  },
  "passwd": {
    "users": [
      {
        "name": "test",
        "uid": 1010,
        "gecos": "user creation test",
        "noCreateHome": true,
        "noUserGroup": true
      }
    ]
  }
}
```

## From Version 1 to 2.0.0

This section will cover the breaking changes made between versions 1 and 2.0.0 of the configuration specification.

### Version

One of the more notable changes was the representation of the config version, moving from an integer to a [Semantic Version][semver] string. Using a Semantic Version will allow the configuration specification to pick up additions and other backward-compatible changes in the future without necessarily requiring the user to update their config. The version number has also moved locations and is now in an Ignition metadata section named "ignition".

The following shows the changes to the version section:

```json ignition
{
  "ignitionVersion": 1
}
```

```json ignition
{
  "ignition": {
    "version": "2.0.0"
  }
}
```

### Files

The `files` section was moved out from under `filesystems` and is now directly under the `storage` section. This was done in order to decouple file definitions from filesystem definitions. This is particularly useful when merging multiple configs together. One config may define a filesystem while another can write files to that filesystem without needing to know the specifics of that filesystem. A common example of this is referencing the universally-defined "root" filesystem which is defined by default inside of Ignition.

The following shows this particular change to the files section:

```json ignition
{
  "storage": {
    "filesystems": [
      {
        "device": "/dev/sdb1",
        "format": "ext4",
        "files": [
          {
            "path": "/foo/bar"
          }
        ]
      }
    ]
  }
}
```

```json ignition
{
  "storage": {
    "filesystems": [
      {
        "name": "example",
        "device": "/dev/sdb1",
        "format": "ext4"
      }
    ],
    "files": [
      {
        "filesystem": "example",
        "path": "/foo/bar"
      }
    ]
  }
}
```

#### Contents

The `contents` section was changed from a simple string to an object. This allows extra properties to be added to file contents (e.g. compression type, content hashs). The source for the file contents has also changed from being inline in the config to a URL. This provides the ability to include the contents inline (via a [data URL][rfc2397]) or to reference a remote resource (via an http URL).

The following shows the changes to the file contents (snipped for clarity):

```json ignition
...

"files": [
  {
    "path": "/foo/bar",
    "contents": "example file\n"
  }
]

...
```

```json ignition
...

"files": [
  {
    "path": "/foo/bar",
    "contents": {
      "source": "data:,example%20file%0A"
    }
  }
]

...
```

#### User and Group

The `uid` and `gid` sections have been moved into new `id` sections under new `user` and `group` sections. This will allow alternate methods of identifying a user or a group (e.g. by name) in the future.

The following shows the changes to the file uid and gid:

```json ignition
...

"files": [
  {
    "path": "/foo/bar",
    "uid": 500,
    "gid": 500
  }
]

...

```

```json ignition
...

"files": [
  {
    "path": "/foo/bar",
    "user": {
      "id": 500
    },
    "group": {
      "id": 500
    }
  }
]

...

```

[semver]: http://semver.org
[rfc2397]: https://tools.ietf.org/html/rfc2397
[operator-notes]: operator-notes.md
[filesystems]: operator-notes.md#filesystem-reuse-semantics
