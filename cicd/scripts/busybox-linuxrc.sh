#!/bin/sh
### BEGIN INIT INFO
# Provides:          arhat
# Required-Start:    $remote_fs networking
# Required-Stop:
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start arhat.
### END INIT INFO

NAME="arhat"
ENVS="HOME=/root TERM=xterm-256color SHELL=/bin/bash"
PIDFILE="/run/${NAME}.pid"

do_start() {
  printf "Starting %s: " "$NAME"
  if start-stop-daemon --start --make-pidfile --pidfile /run/"$NAME".pid --background --exec /bin/sh -- -c "$ENVS /usr/bin/arhat -c /etc/arhat/config.yaml" ; then
    printf "ok."
  else
    printf "failed"
  fi
  printf "\n"
}

do_stop() {
  if [ -f "$PIDFILE" ]; then
    printf "Stopping %s: " "$NAME"
    if start-stop-daemon --stop --pidfile "$PIDFILE"; then
      printf "ok.\n"
    else
      printf "failed.\n"
    fi
  else
    printf "%s is not running.\n" "$NAME"
  fi
}

case "$1" in
start)
  do_start
  ;;
stop)
  do_stop
  ;;
restart)
  do_stop

  sleep 1

  do_start
  ;;
*)
  echo "Usage: $NAME {start|stop|restart}" >&2
  exit 1
  ;;
esac

exit 0
