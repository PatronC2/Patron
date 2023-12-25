#!/bin/bash

if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root" 
   exit 1
fi

function show_help {
    echo "Usage: $0 [-d ] [-w ] [ -s <your_ip_address> ]"
    echo "Options:"
    echo "  -d    Use default options"
    echo "  -w    Wipe Database"
    echo "  -s    <your_ip_address>   Server Ip address"
    echo "  -p    Prompts you to enter passwords"
    echo "  -h    Show this help message"
}

function set_global_default_variable {
    webserverip="0.0.0.0"
    webserverport="8000"
    reactclientip="0.0.0.0"
    reactclientport="8081"
    c2serverip=""
    c2serverport="9000"
    dockerinternal="172.18.0"
    nginxip=""
    nginxport="8082"
    bottoken=""
    dbhost="172.18.0.9"
    dbuser="patron"
    dbport="5432"
    dbname="patron"
}

function ask_prompt {
   
   echo "Note: Webserver, ReactClient, DB and C2 server can't be on the same port and must be an unused port"
   read -p "Enter APISERVER IP: " webserverip
   read -p "Enter APISERVER PORT: " webserverport
   read -p "Enter REACTCLIENT IP: " reactclientip
   read -p "Enter REACTCLIENT PORT: " reactclientport
   echo "Note: To listen on all inteface, leave C2SERVER IP blank"
   read -p "Enter C2SERVER IP: " c2serverip
   read -p "Enter C2SERVER PORT: " c2serverport
   read -p "Enter DOCKER INTERNAL NETWORK e.g 172.18.0 (without last octect): " dockerinternal
   # read -p "Enter NGINX IP (exposed ip): " nginxip
   read -p "Enter NGINX PORT: " nginxport
   read -p "Enter Database Host: " dbhost
   read -p "Enter Database Port: " dbport
   read -p "Enter Database User: " dbuser
   read -p "Enter Database Name: " dbname
   read -p "Enter DISCORD BOT TOKEN: " bottoken
}

function wipe_db {
   rm -rf data/postgres_data
   rm -rf nginx/.htpasswd
   echo "Database Wiped!"
   echo "htpasswd Wiped!"
   dbpass=`openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12`
   basicpass=`openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12`
   set_htpasswd
}

function pass_prompt {
   read -p "Enter Database Password: " dbpass
   read -p "Enter Basic Auth Password: " basicpass
   set_htpasswd
}

function set_htpasswd {
   # Check if the .htpasswd file exists
   if [ ! -f "nginx/.htpasswd" ]; then
      # Create the file for the first time
      htpasswd -B -b -c "nginx/.htpasswd" "$dbuser" "$basicpass"
      echo "Password for user $dbuser has been added to nginx/.htpasswd"
   else
      # Append the user to the existing file
      htpasswd -B -b "nginx/.htpasswd" "$dbuser" "$basicpass"
      echo "Password for user $dbuser has been added to nginx/.htpasswd"
   fi
}

function prereq_app_check {
   base64=$(which base64)
   openssl=$(which openssl)
   npm=$(which npm)
   go=$(which go)
   htpass=$(which htpasswd)

   # Prereqs
   #base64 check
   if [ -x "$base64" ]; then
   echo "base64 Check Ok"
   else
   echo "Install base64"
   exit
   fi

   #openssl check
   if [ -x "$openssl" ]; then
   echo "openssl Check Ok"
   else
   echo "Install openssl"
   exit
   fi

   #npm check
   if [ -x "$npm" ]; then
   echo "npm Check Ok"
   else
   echo "Install npm: sudo apt install npm nodejs"
   exit
   fi

   #go check
   if [ -x "$go" ]; then
   echo "go Check Ok"
   else
   echo "Install go: sudo apt install golang"
   exit
   fi

   #go check
   if [ -x "$htpass" ]; then
   echo "htpasswd Check Ok"
   else
   echo "Install htpasswd: sudo apt install htpasswd"
   exit
   fi
}

default=""
clean_db=""
ipaddress=""
postgres_pass=""
# Parse command line arguments using getopts
while getopts "dws:ph" opt; do
    case $opt in
        d)
            set_global_default_variable
            default="y"
            ;;
        w)
            wipe_db
            clean_db="y"
            ;;
        s)
            ipaddress="$OPTARG"
            ;;
        p)
            pass_prompt
            postgres_pass="y"
            ;;
        h)
            show_help
            exit 0
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            show_help
            exit 1
            ;;
    esac
done

if [ -z "$ipaddress" ]; then
   echo "Error: Set your ip with -s"
   show_help
   exit 1
else
   nginxip=$ipaddress
fi

if [ -z "$default" ]; then
   ask_prompt
fi

# Check if both -w and -p flags are provided together
if [[ -n "$postgres_pass" && -n "$clean_db" ]] || [[ -z "$postgres_pass" && -z "$clean_db" ]]; then
    echo "Error: Both -w and -p flags must be used separately. Or at least one must be used"
    show_help
    exit 1
fi

# Shift the processed options
shift $((OPTIND-1))

# # Validate the number of arguments
# if [ "$#" -lt 1 ]; then
#     echo "Error: At least one input file is required."
#     show_help
#     exit 1
# fi

prereq_app_check


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
rm -rf .env
rm -rf Web/client/.env

encpubkey=`base64 -w 0 certs/server.pem`

# server env
echo "WEBSERVER_IP=$webserverip" >> .env
echo "WEBSERVER_PORT=$webserverport" >> .env
echo "C2SERVER_IP=$c2serverip" >> .env
echo "C2SERVER_PORT=$c2serverport" >> .env
echo "PUBLIC_KEY=$encpubkey" >> .env
echo "BOT_TOKEN=$bottoken" >> .env
echo "DB_HOST=$dbhost" >> .env
echo "DB_PORT=$dbport" >> .env
echo "DB_USER=$dbuser" >> .env
echo "DB_PASS=$dbpass" >> .env
echo "DB_NAME=$dbname" >> .env
echo "DOCKER_INTERNAL=$dockerinternal" >> .env
echo "REACT_APP_NGINX_PORT=$nginxport" >> .env
echo "REACT_APP_NGINX_IP=$nginxip" >> .env
echo "REACT_SERVER_IP=$reactclientip" >> .env
echo "REACT_SERVER_PORT=$reactclientport" >> .env
echo "BASIC_AUTH_PASS=$basicpass" >> .env

#webclient env
echo "REACT_APP_WEBSERVER_IP=$webserverip" >> Web/client/.env
echo "REACT_APP_WEBSERVER_PORT=$webserverport" >> Web/client/.env
echo "REACT_APP_NGINX_IP=$nginxip" >> Web/client/.env
echo "REACT_APP_NGINX_PORT=$nginxport" >> Web/client/.env
echo "HOST=$reactclientip" >> Web/client/.env
echo "PORT=$reactclientport" >> Web/client/.env


#go mod tidy
echo "Running: Go mod tidy"
go mod tidy

# npm install
echo "Installing node modules..."

cd Web/client && npm install && cd ../../


echo ""
echo ""
echo "------------------------------------------Raw Dog Run------------------------------------------"
echo ""
echo "Run 'sudo go run server/server.go' to start the C2 server"
echo ""
echo "Run 'sudo go run Web/server/webserver.go' to start the api sever"
echo ""
echo "Run 'cd Web/client && npm start' to start start the web client"
echo ""
echo "Run 'sudo go run bot/bot.go' to start the Discord Bot if the DISCORD BOT_TOKEN Was provided"
echo ""
echo ""
echo ""
echo "------------------------------------------ Docker ------------------------------------------"
echo ""
echo "Spin up: Run 'docker compose up --remove-orphans'"
echo ""
echo "Tear down: Run 'docker compose down'"
echo ""
echo ""
echo "------------------------------------------ Informational --------------------------------------"
echo ""
echo "Visit http://$nginxip:$nginxport for Web"
echo ""
echo "C2 Server on $nginxip:$c2serverport"
echo ""
echo "See .env and Web/client/.env to tweak enviroment variables (not advised)"
echo ""

