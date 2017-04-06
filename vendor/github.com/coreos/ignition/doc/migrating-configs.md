# Migrating Between Configuration Versions

Occasionally, there are changes made to Ignition's configuration that break backward compatibility. While this is not a concern for running machines (since Ignition only runs one time during first boot), it is a concern for those who maintain configuration files. This document serves to detail each of the breaking changes and tries to provide some reasoning for the change. This does not cover all of the changes to the spec - just those that need to be considered when migrating from one version to the next.

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
