#!/bin/bash
# Script to be ran by systemd, please do not manually run this script
: ${SERVER_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

cd $SERVER_DIR
/usr/bin/go run server/server.go >> /var/log/patron/server.log 2>&1 & disown
sleep 10
/usr/bin/go run Web/server/webserver.go >> /var/log/patron/webserver.log 2>&1 & disown
sleep 5
cd $SERVER_DIR/Web/client
/usr/bin/npm start >> /var/log/patron/webclient.log 2>&1
