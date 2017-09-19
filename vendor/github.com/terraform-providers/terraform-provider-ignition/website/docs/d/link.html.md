---
layout: "ignition"
page_title: "Ignition: ignition_link"
sidebar_current: "docs-ignition-datasource-link"
description: |-
  Describes a link to be created in a particular filesystem.
---

# ignition\_link

Describes a link to be created in a particular filesystem.

## Example Usage

```hcl
data "ignition_link" "symlink" {
	filesystem = "foo"
	path = "/symlink"
    target = "/foo"
}
```

## Argument Reference

The following arguments are supported:

* `filesystem` - (Required) The internal identifier of the filesystem. This matches the last filesystem with the given identifier. This should be a valid name from a _ignition\_filesystem_ resource.

* `path` - (Required) The absolute path to the link.

* `target` - (Required) The target path of the link.

* `hard` - (Optional) A symbolic link is created if this is false, a hard one if this is true.

* `uid` - (Optional) The user ID of the owner.

* `gid` - (Optional) The group ID of the owner.

## Attributes Reference

The following attributes are exported:

* `id` - ID used to reference this resource in _ignition_config_.
