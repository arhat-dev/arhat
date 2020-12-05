#!/bin/sh
# shellcheck disable=SC2034

# PROVIDE: arhat
# REQUIRE: networking
# KEYWORD: shutdown

. /etc/rc.subr

name="arhat"
desc="The reference EdgeDevice agent"
rcvar="arhat_enable"
command="/usr/local/bin/${name}"
pidfile="/var/run/${name}.pid"

# Set defaults
: ${arhat_enable:="NO"}
: ${arhat_user:="root"}
: ${arhat_config:="/usr/local/etc/arhat/config.yaml"}
: ${arhat_flags:="-c $arhat_config"}

load_rc_config $name
run_rc_command "$1"
