---
layout: "ignition"
page_title: "Ignition: ignition_directory"
sidebar_current: "docs-ignition-datasource-directory"
description: |-
  Describes a directory to be created in a particular filesystem.
---

# ignition\_directory

Describes a directory to be created in a particular filesystem.

## Example Usage

```hcl
data "ignition_directory" "folder" {
	filesystem = "foo"
	path = "/folder"
}
```

## Argument Reference

The following arguments are supported:

* `filesystem` - (Required) The internal identifier of the filesystem. This matches the last filesystem with the given identifier. This should be a valid name from a _ignition\_filesystem_ resource.

* `path` - (Required) The absolute path to the directory.

* `mode` - (Optional) The directory's permission mode. Note that the mode must be properly specified as a decimal value (i.e. 0755 -> 493).

* `uid` - (Optional) The user ID of the owner.

* `gid` - (Optional) The group ID of the owner.

## Attributes Reference

The following attributes are exported:

* `id` - ID used to reference this resource in _ignition_config_.
