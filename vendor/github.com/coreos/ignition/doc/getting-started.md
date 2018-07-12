# Getting Started with Ignition

*Ignition* is a low-level system configuration utility. The Ignition executable is part of the temporary initial root filesystem, the *initramfs*. When Ignition runs, it finds configuration data in a named location for a given environment, such as a file or URL, and applies it to the machine before `switch_root` is called to pivot to the machine's root filesystem.

Ignition uses a JSON configuration file to represent the set of changes to be made. The format of this config is detailed [in the specification][configspec] and the [MIME type][mime] is registered with IANA. One of the most important parts of this config is the version number. This **must** match the version number accepted by Ignition. If the config version isn't accepted by Ignition, Ignition will fail to run and the machine will not boot. This can be seen by inspecting the console output of the failed machine. For more information, check out the [troubleshooting section][troubleshooting].

## Providing a Config

Ignition will choose where to look for configuration based on the underlying platform. A list of [supported platforms][platforms] and metadata sources is provided for reference.

The configuration must be passed to Ignition through the designated data source. Please refer to Ignition [config examples][examples] to learn about writing config files. The provided configuration will be appended to the universal base configuration:

```json ignition
...
"storage": {
  "filesystems": [{
    "name": "root",
    "path": "/sysroot"
  }]
}
...
```

This data source can be overriden by specifying a configuration URL via the kernel command-line options.

## Troubleshooting

### Gathering Logs

The single most useful piece of information needed when troubleshooting is the log from Ignition. Ignition runs in multiple stages so it's easiest to filter by the syslog identifier: `ignition`. When using systemd, this can be accomplished with the following command:

```
journalctl --identifier=ignition --all
```

In the event that this doesn't yield any results, running as root may help. There are circumstances where the journal isn't owned by the systemd-journal group or the current user is not a part of that group.

### Increasing Verbosity

In cases where the machine fails to boot, it's sometimes helpful to ask journald to log more information to the console. This makes it easy to access the Ignition logs in environments where no interactive console is available. The following kernel parameter will increase the console's log output, making all of Ignition's logs visible:

`systemd.journald.max_level_console=debug`

### Validating the Configuration

One common cause for Ignition failures is a malformed configuration (e.g. a misspelled section or incorrect hierarchy). Ignition will log errors, warnings, and other notes about the configuration that it parsed, so this can be used to debug issues with the configuration provided. As a convenience, CoreOS hosts an [online validator][validator] which can be used to quickly verify configurations.

### Enabling systemd Services

When Ignition enables systemd services, it doesn't directly create the symlinks necessary for systemd; it leverages [systemd presets][preset]. Presets are only evaluated on [first-boot][conditions], which can result in confusion if Ignition is forced to run more than once. Any systemd services which have been enabled in the configuration after the first boot won't actually be enabled after the next invocation of Ignition. `systemctl preset-all` will need to be manually invoked to create the necessary symlinks, enabling the services.

Ignition is not typically run more than once during a machine's lifetime in a given role, so this situation requiring manual systemd intervention does not commonly arise.

[conditions]: https://www.freedesktop.org/software/systemd/man/systemd.unit.html#ConditionArchitecture=
[configspec]: configuration-v2_0.md
[examples]: examples.md
[mime]: http://www.iana.org/assignments/media-types/application/vnd.coreos.ignition+json
[platforms]: supported-platforms.md
[preset]: https://www.freedesktop.org/software/systemd/man/systemd.preset.html
[troubleshooting]: #troubleshooting
[validator]: https://coreos.com/validate
