#!/bin/sh /etc/rc.common
# shellcheck disable=SC2034

# procd init script for arhat on openwrt ...

USE_PROCD=1

START=95
STOP=01

BIN="/usr/bin/arhat"
CONFIG="/etc/config/arhat.yaml"

start_service() {
  procd_open_instance

  procd_set_param command ${BIN} -c ${CONFIG}

  procd_set_param user root
  procd_set_param env HOME=/root TERM=xterm SHELL=/bin/ash
  procd_set_param respawn

  procd_close_instance
}

stop_service() {
  name="$(basename "${BIN}")"
  kill "$(pidof "${name}")"
}
