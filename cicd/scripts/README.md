# Scripts for arhat startup

## Usage

Depending on your system's init system, you should choose one script from following

### `busybox-linuxrc-arhat.sh`

[busybox linuxrc](https://github.com/mirror/busybox/blob/9b4a9d96b89f06355ad9551d782d34506699aac8/init/init.c#L18-L27)

Target Location: `/etc/init.d/arhat`
Auto start on boot: `update-rc.d arhat enable 2` (runlevel 2)

### `openrc-arhat.sh`

[OpenRC](https://wiki.gentoo.org/wiki/OpenRC) is the init system used by Gentoo by default

Target Location: `/etc/init.d/arhat`
Auto start on boot: `rc-update add arhat`

### `procd-arhat.sh`

[Procd](https://openwrt.org/docs/techref/procd) is the init system used by OpenWRT

Target Location: `/etc/init.d/arhat`
Auto start on boot: `/etc/init.d/arhat enable`

### `systemd-arhat.service`

[systemd](https://wiki.archlinux.org/index.php/systemd) is a widely used init system for linux

Target Location: `/lib/systemd/system/arhat.service` or `/etc/systemd/system/arhat.service`
Auto start on boot: `systemctl daemon-reload && systemctl enable arhat`
