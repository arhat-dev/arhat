# Startup Scripts for `arhat`

## Usage

Depending on your system's init system, you should choose one script from following

| Init System                        | Used by            | Startup Script                               | Target Location                                                                                                                                   | Start on boot                                       |
| ---------------------------------- | ------------------ | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| [BSD init][bsd-init]               | FreeBSD            | [`bsd-init.sh`](./bsd-init.sh)               | `/usr/local/etc/rc.d/arhat`                                                                                                                       | Add `enable_arhat=yes` to `/usr/local/etc/rc.conf`  |
| [busybox linuxrc][busybox-linuxrc] | Alpine             | [`busybox-linuxrc.sh`](./busybox-linuxrc.sh) | `/etc/init.d/arhat`                                                                                                                               | `update-rc.d arhat enable 2` (runlevel 2)           |
| [launchd][launchd]                 | macOS              | [`launchd.plist`](./launchd.plist)           | `/Users/${USERNAME}/Library/LaunchAgents/dev.arhat.arhat.plist` (unprivileged user) or `/Library/LaunchDaemons/dev.arhat.arhat.plist` (root user) | already set with `AutoLoad` tag                     |
| [OpenRC][openrc]                   | Gentoo             | [`openrc.sh`](./openrc.sh)                   | `/etc/init.d/arhat`                                                                                                                               | `rc-update add arhat`                               |
| [Procd][procd]                     | OpenWRT            | [`procd.sh`](./procd.sh)                     | `/etc/init.d/arhat`                                                                                                                               | `/etc/init.d/arhat enable`                          |
| [Systemd][systemd]                 | many linux distros | [`systemd.service`](./systemd.service)       | `/lib/systemd/system/arhat.service` or `/etc/systemd/system/arhat.service`                                                                        | `systemctl daemon-reload && systemctl enable arhat` |

[busybox-linuxrc]: https://github.com/mirror/busybox/blob/9b4a9d96b89f06355ad9551d782d34506699aac8/init/init.c#L18-L27
[bsd-init]: https://www.freebsd.org/cgi/man.cgi?query=init
[launchd]: https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
[openrc]: https://wiki.gentoo.org/wiki/OpenRC
[procd]: https://openwrt.org/docs/techref/procd
[systemd]: https://wiki.archlinux.org/index.php/systemd

__NOTE:__ You have to deploy `arhat` and its configuration according to the startup script
