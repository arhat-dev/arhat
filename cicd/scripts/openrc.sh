#!/sbin/openrc-run
# shellcheck disable=SC2034

# openrc init script for arhat on gentoo ...

export TERM=xterm
# export HOME=/home/pi
# export SHELL=/usr/bin/zsh

command="/usr/local/bin/arhat"
command_args="-c /etc/arhat/config.yaml"
command_background="yes"

pidfile="/run/arhat.pid"

depend() {
  need net
}

start_pre() {
  # ensure we have netfilter kernel module setup
  modprobe iptable_mangle
  modprobe iptable_nat
  modprobe ip6table_mangle
  modprobe ip6table_nat

  # turn on forwarding for all kinds of network traffic
  sysctl -w net.ipv4.ip_forward=1
  sysctl -w net.ipv6.conf.all.forwarding=1

  # disable bridge netfilter for udp tproxy
  modprobe br_netfilter
  sysctl -w net.bridge.bridge-nf-call-iptables=0
  sysctl -w net.bridge.bridge-nf-call-ip6tables=0
}
