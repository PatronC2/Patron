#!/bin/bash

base64=$(which base64)
openssl=$(which openssl)
npm=$(which npm)
npm=$(which go)
dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
dirsedsafe=$(printf '%s\n' "$dir" | sed -e 's/[]\/$*.^[]/\\&/g');

#base64 check
if [ -x $base64 ]; then
echo "base64 Check Ok"
else
echo "Install base64"
exit
fi

#openssl check
if [ -x $openssl ]; then
echo "openssl Check Ok"
else
echo "Install openssl"
exit
fi

#npm check
if [ -x $npm ]; then
echo "npm Check Ok"
else
echo "Install npm: sudo apt install npm nodejs"
exit
fi

#go check
if [ -x $go ]; then
echo "go Check Ok"
else
echo "Install go: sudo apt install golang"
exit
fi

#install certs
echo "Generating certs"
[ ! -d "$PWD/certs" ] && mkdir certs
rm -rf certs/server.key
rm -rf certs/server.pem
openssl ecparam -genkey -name prime256v1 -out certs/server.key
openssl req -new -x509 -key certs/server.key -out certs/server.pem -days 3650 -subj "/C=US/ST=Maryland/L=Towson/O=Case Studies/OU=Offensive Op/CN=example.com"

#Setting up agents dir
echo "Setting up agents dir"
[ ! -d "$PWD/agents" ] && mkdir agents

# Set Env file
echo "Setting environment variables"
echo "Note: Webserver and C2 server can't be on the same port and must be an unused port"
rm -rf .env
rm -rf Web/client/.env
touch .env
read -p "Enter WEBSERVER IP: " webserverip
read -p "Enter WEBSERVER PORT: " webserverport
echo "Note: To listen on all inteface, leave C2SERVER IP blank"
read -p "Enter C2SERVER IP: " c2serverip
read -p "Enter C2SERVER PORT: " c2serverport
echo "Note: Leave discord bot token blank if you don't want"
read -p "Enter DISCORD BOT TOKEN: " bottoken
encpubkey=`base64 -w 0 certs/server.pem`

# server env
echo "WEBSERVER_IP=$webserverip" >> .env
echo "WEBSERVER_PORT=$webserverport" >> .env
echo "C2SERVER_IP=$c2serverip" >> .env
echo "C2SERVER_PORT=$c2serverport" >> .env
echo "PUBLIC_KEY=$encpubkey" >> .env
echo "BOT_TOKEN=$bottoken" >> .env

#webclient env
echo "REACT_APP_WEBSERVER_IP=$webserverip" >> Web/client/.env
echo "REACT_APP_WEBSERVER_PORT=$webserverport" >> Web/client/.env
echo "HOST=$webserverip" >> Web/client/.env
echo "PORT=$webserverport" >> Web/client/.env

read -p "Do you want to reset the database (this will clear any keylogs) (y/n): " resetchoice

if [ "$resetchoice" = 'y' ]; then
rm -rf data/sqlite-database.db
echo "Database Wiped!"
else
echo "Good Choice"
fi

# configure patron service
mkdir /var/log/patron
sed -i "s/SCRIPT_FILE/$dirsedsafe/g" $dir/patron.service
cp $dir/patron.service /etc/init.d/patron
chmod 755 /etc/init.d/patron
git -C $dir restore patron.service
systemctl enable patron

#go mod tidy
echo "Go mod tidy"
go mod tidy

# npm install

echo "Installing node modules..."

cd Web/client && npm install && cd ../../ 
echo ""
echo ""
echo ""
echo ""
echo "    Use Compiled binary      "
echo "Run './build/server' to start the C2 Server"
echo "Run './build/webserver' to start the Web Server"
echo "Run 'cd Web/client && npm start' to start the Web Client"
echo ""
echo ""
echo "            OR               "
echo ""
echo ""
echo "       Go installed?         "
echo "Run 'go run server/server.go' to start the C2 Server"
echo "Run 'go run Web/server/webserver.go'"
echo "Run 'cd Web/client && npm start' to start the Web Client"
