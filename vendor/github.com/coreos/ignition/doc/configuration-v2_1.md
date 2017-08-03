# Configuration Specification v2.1.0 #

The Ignition configuration is a JSON document conforming to the following specification, with **_italicized_** entries being optional:

* **ignition** (object): metadata about the configuration itself.
  * **version** (string): the semantic version number of the spec. The spec version must be compatible with the latest version (`2.0.0`). Compatibility requires the major versions to match and the spec version be less than or equal to the latest version.
  * **_config_** (objects): options related to the configuration.
    * **_append_** (list of objects): a list of the configs to be appended to the current config.
      * **source** (string): the URL of the config. Supported schemes are http, https, s3, and tftp. Note: When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
      * **_verification_** (object): options related to the verification of the config.
        * **_hash_** (string): the hash of the config, in the form `<type>-<value>` where type is sha512.
    * **_replace_** (object): the config that will replace the current.
      * **source** (string): the URL of the config. Supported schemes are http, https, s3, and tftp. Note: When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
      * **_verification_** (object): options related to the verification of the config.
        * **_hash_** (string): the hash of the config, in the form `<type>-<value>` where type is sha512.
  * **_timeouts_** (object): options relating to http timeouts when fetching files over http or https.
    * **_httpResponseHeaders_** (integer) the time to wait (in seconds) for the server's repsonse headers (but not the body) after making a request. 0 indicates no timeout. Default is 10 seconds.
    * **_httpTotal_** (integer) the time limit (in seconds) for the operation (connection, request, and response), including retries. 0 indicates no timeout. Default is 0.
* **_storage_** (object): describes the desired state of the system's storage devices.
  * **_disks_** (list of objects): the list of disks to be configured and their options.
    * **device** (string): the absolute path to the device. Devices are typically referenced by the `/dev/disk/by-*` symlinks.
    * **_wipeTable_** (boolean): whether or not the partition tables shall be wiped. When true, the partition tables are erased before any further manipulation. Otherwise, the existing entries are left intact.
    * **_partitions_** (list of objects): the list of partitions and their configuration for this particular disk.
      * **_label_** (string): the PARTLABEL for the partition.
      * **_number_** (integer): the partition number, which dictates it's position in the partition table (one-indexed). If zero, use the next available partition slot.
      * **_size_** (integer): the size of the partition (in sectors). If zero, the partition will fill the remainder of the disk.
      * **_start_** (integer): the start of the partition (in sectors). If zero, the partition will be positioned at the earliest available part of the disk.
      * **_typeGuid_** (string): the GPT [partition type GUID][part-types]. If omitted, the default will be 0FC63DAF-8483-4772-8E79-3D69D8477DE4 (Linux filesystem data).
      * **_guid_** (string): the GPT unique partition GUID.
  * **_raid_** (list of objects): the list of RAID arrays to be configured.
    * **name** (string): the name to use for the resulting md device.
    * **level** (string): the redundancy level of the array (e.g. linear, raid1, raid5, etc.).
    * **devices** (list of strings): the list of devices (referenced by their absolute path) in the array.
    * **_spares_** (integer): the number of spares (if applicable) in the array.
  * **_filesystems_** (list of objects): the list of filesystems to be configured and/or used in the "files" section. Either "mount" or "path" needs to be specified.
    * **_name_** (string): the identifier for the filesystem, internal to Ignition. This is only required if the filesystem needs to be referenced in the "files" section.
    * **_mount_** (object): contains the set of mount and formatting options for the filesystem. A non-null entry indicates that the filesystem should be mounted before it is used by Ignition.
      * **device** (string): the absolute path to the device. Devices are typically referenced by the `/dev/disk/by-*` symlinks.
      * **format** (string): the filesystem format (ext4, btrfs, xfs, vfat, or swap).
      * **_wipeFilesystem_** (boolean): whether or not to wipe the device before filesystem creation, see [the documentation on filesystems](operator-notes.md#filesystem-reuse-semantics) for more information.
      * **_label_** (string): the label of the filesystem.
      * **_uuid_** (string): the uuid of the filesystem.
      * **_options_** (list of strings): any additional options to be passed to the format-specific mkfs utility.
      * **_create_** (object, DEPRECATED): contains the set of options to be used when creating the filesystem.
        * **_force_** (boolean, DEPRECATED): whether or not the create operation shall overwrite an existing filesystem.
        * **_options_** (list of strings, DEPRECATED): any additional options to be passed to the format-specific mkfs utility.
    * **_path_** (string): the mount-point of the filesystem. A non-null entry indicates that the filesystem has already been mounted by the system at the specified path. This is really only useful for "/sysroot".
  * **_files_** (list of objects): the list of files to be written.
    * **filesystem** (string): the internal identifier of the filesystem in which to write the file. This matches the last filesystem with the given identifier.
    * **path** (string): the absolute path to the file.
    * **_contents_** (object): options related to the contents of the file.
      * **_compression_** (string): the type of compression used on the contents (null or gzip). Compression cannot be used with S3.
      * **_source_** (string): the URL of the file contents. Supported schemes are http, https, tftp, s3, and [data][rfc2397]. When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
      * **_verification_** (object): options related to the verification of the file contents.
        * **_hash_** (string): the hash of the config, in the form `<type>-<value>` where type is sha512.
    * **_mode_** (integer): the file's permission mode. Note that the mode must be properly specified as a **decimal** value (i.e. 0644 -> 420).
    * **_user_** (object): specifies the file's owner.
      * **_id_** (integer): the user ID of the owner.
      * **_name_** (string): the user name of the owner.
    * **_group_** (object): specifies the group of the owner.
      * **_id_** (integer): the group ID of the owner.
      * **_name_** (string): the group name of the owner.
  * **_directories_** (list of objects): the list of directories to be created.
    * **filesystem** (string): the internal identifier of the filesystem in which to create the directory. This matches the last filesystem with the given identifier.
    * **path** (string): the absolute path to the directory.
    * **_mode_** (integer): the directory's permission mode. Note that the mode must be properly specified as a **decimal** value (i.e. 0755 -> 493).
    * **_user_** (object): specifies the directory's owner.
      * **_id_** (integer): the user ID of the owner.
      * **_name_** (string): the user name of the owner.
    * **_group_** (object): specifies the group of the owner.
      * **_id_** (integer): the group ID of the owner.
      * **_name_** (string): the group name of the owner.
  * **_links_** (list of objects): the list of links to be created
    * **filesystem** (string): the internal identifier of the filesystem in which to write the link. This matches the last filesystem with the given identifier.
    * **path** (string): the absolute path to the link
    * **_user_** (object): specifies the symbolic links's owner.
      * **_id_** (integer): the user ID of the owner.
      * **_name_** (string): the user name of the owner.
    * **_group_** (object): specifies the group of the owner.
      * **_id_** (integer): the group ID of the owner.
      * **_name_** (string): the group name of the owner.
    * **target** (string): the target path of the link
    * **_hard_** (boolean): a symbolic link is created if this is false, a hard one if this is true.
* **_systemd_** (object): describes the desired state of the systemd units.
  * **_units_** (list of objects): the list of systemd units.
    * **name** (string): the name of the unit. This must be suffixed with a valid unit type (e.g. "thing.service").
    * **_enable_** (boolean, DEPRECATED): whether or not the service shall be enabled. When true, the service is enabled. In order for this to have any effect, the unit must have an install section.
    * **_enabled_** (boolean): whether or not the service shall be enabled. When true, the service is enabled. When false, the service is disabled. When omitted, the service is unmodified. In order for this to have any effect, the unit must have an install section.
    * **_mask_** (boolean): whether or not the service shall be masked. When true, the service is masked by symlinking it to `/dev/null`.
    * **_contents_** (string): the contents of the unit.
    * **_dropins_** (list of objects): the list of drop-ins for the unit.
      * **name** (string): the name of the drop-in. This must be suffixed with ".conf".
      * **_contents_** (string): the contents of the drop-in.
* **_networkd_** (object): describes the desired state of the networkd files.
  * **_units_** (list of objects): the list of networkd files.
    * **name** (string): the name of the file. This must be suffixed with a valid unit type (e.g. "00-eth0.network").
    * **_contents_** (string): the contents of the networkd file.
* **_passwd_** (object): describes the desired additions to the passwd database.
  * **_users_** (list of objects): the list of accounts that shall exist.
    * **name** (string): the username for the account.
    * **_passwordHash_** (string): the encrypted password for the account.
    * **_sshAuthorizedKeys_** (list of strings): a list of SSH keys to be added to the user's authorized_keys.
    * **_uid_** (integer): the user ID of the account.
    * **_gecos_** (string): the GECOS field of the account.
    * **_homeDir_** (string): the home directory of the account.
    * **_noCreateHome_** (boolean): whether or not to create the user's home directory. This only has an effect if the account doesn't exist yet.
    * **_primaryGroup_** (string): the name of the primary group of the account.
    * **_groups_** (list of strings): the list of supplementary groups of the account.
    * **_noUserGroup_** (boolean): whether or not to create a group with the same name as the user. This only has an effect if the account doesn't exist yet.
    * **_noLogInit_** (boolean): whether or not to add the user to the lastlog and faillog databases. This only has an effect if the account doesn't exist yet.
    * **_shell_** (string): the login shell of the new account.
    * **_system_** (bool): whether or not to make the account a system account. This only has an effect if the account doesn't exist yet.
    * **_create_** (object, DEPRECATED): contains the set of options to be used when creating the user. A non-null entry indicates that the user account shall be created. This object has been marked for deprecation, please use the **_users_** level fields instead.
      * **_uid_** (integer): the user ID of the new account.
      * **_gecos_** (string): the GECOS field of the new account.
      * **_homeDir_** (string): the home directory of the new account.
      * **_noCreateHome_** (boolean): whether or not to create the user's home directory.
      * **_primaryGroup_** (string): the name or ID of the primary group of the new account.
      * **_groups_** (list of strings): the list of supplementary groups of the new account.
      * **_noUserGroup_** (boolean): whether or not to create a group with the same name as the user.
      * **_noLogInit_** (boolean): whether or not to add the user to the lastlog and faillog databases.
      * **_shell_** (string): the login shell of the new account.
  * **_groups_** (list of objects): the list of groups to be added.
    * **name** (string): the name of the group.
    * **_gid_** (integer): the group ID of the new group.
    * **_passwordHash_** (string): the encrypted password of the new group.

[part-types]: http://en.wikipedia.org/wiki/GUID_Partition_Table#Partition_type_GUIDs
[rfc2397]: https://tools.ietf.org/html/rfc2397
