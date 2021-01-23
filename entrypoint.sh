#!/bin/sh

PIDFILE=/root/.local/share/heedy/heedy.pid
if test -f "$PIDFILE"; then
    echo "PID file $PIDFILE exists."
    echo "We assume no other instance is running in another container and delete it."
    rm $PIDFILE
fi


# Execute CMD
exec "${@}"
