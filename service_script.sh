#!/bin/bash
# Script to be ran by systemd, please do not manually run this script
: ${SERVER_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

/$SERVER_DIR/build/server >> /var/log/patron/server.log 2>&1 & disown
/$SERVER_DIR/build/webserver >> /var/log/patron/webserver.log 2>&1 & disown
cd /$SERVER_DIR/Web/client && npm start >> /var/log/patron/webclient.log 2>&1 & disown
