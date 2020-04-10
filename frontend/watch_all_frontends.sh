#!/bin/bash
# The frontend of heedy isn't just this folder - it also includes all of the builtin plugins.
# This script sets up watching for all of the builtin plugins' frontends.

# https://unix.stackexchange.com/questions/107400/ctrl-c-with-two-simultaneous-commands-in-bash
trap killall SIGINT
killall() {
    kill 0
}
# https://stackoverflow.com/questions/3349105/how-to-set-current-working-directory-to-the-directory-of-the-script
cd "${0%/*}"

(cd ../plugins/dashboard/frontend;npm run debug) &
(cd ../plugins/timeseries/frontend;npm run debug) &
(cd ../plugins/notifications/frontend;npm run debug) &
(cd ../plugins/registry/frontend;npm run debug) &
npm run debug