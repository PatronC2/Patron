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
    nginxport="8443"
    bottoken=""
    dbhost="patron_c2_postgres"
    dbuser="patron"
    dbport="5432"
    dbname="patron"
    patronUsername="patron"
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
   read -p "Enter NGINX PORT: " nginxport
   read -p "Enter Database Host: " dbhost
   read -p "Enter Database Port: " dbport
   read -p "Enter Database User: " dbuser
   read -p "Enter Database Name: " dbname
   read -p "Enter DISCORD BOT TOKEN: " bottoken
}

function wipe_db {
   rm -rf data/postgres_data
   echo "Database Wiped!"
   dbpass=`openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12`
   patronPassword=`openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12`
}

function pass_prompt {
   read -p "Enter Database Password: " dbpass
   read -p "Enter UI Username: " patronUsername
   read -p "Enter UI Password: " patronPassword
}

function prereq_app_check {
   base64=$(which base64)
   openssl=$(which openssl)
   docker=$(which docker)

   # Prereqs
   # base64 check
   if [ -x "$base64" ]; then
      echo "base64 Check Ok"
   else
      echo "Install base64"
      exit 1
   fi

   # openssl check
   if [ -x "$openssl" ]; then
      echo "openssl Check Ok"
   else
      echo "Install openssl"
      exit 1
   fi

   # docker check
   if [ -x "$docker" ]; then
      echo "Docker Check Ok"
   else
      echo "Docker is not installed. Please install Docker."
      exit 1
   fi

   # Check if Docker is running
   if ! docker info > /dev/null 2>&1; then
      echo "Docker daemon is not running. Please start Docker."
      exit 1
   else
      echo "Docker is running"
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
rm -rf ui/.env

encpubkey=`base64 -w 0 certs/server.pem`
# Generate JWT key for API auth
JWT_KEY=$(openssl rand -base64 32)

# Put the pwd into the .env, this is needed for sharing docker volumes for the API
REPO_DIR=`pwd`

# server env
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
echo "ADMIN_AUTH_USER=$patronUsername" >> .env
echo "ADMIN_AUTH_PASS=$patronPassword" >> .env
echo "JWT_KEY=$JWT_KEY" >> .env
# Need full path to the repo for the API building payloads
echo  "REPO_DIR=$REPO_DIR" >> .env
# We cannot change the name of these, locked in by node
echo "HOST=$reactclientip" >> .env
echo "PORT=$reactclientport" >> .env

# make log dir
mkdir -p logs

echo ""
echo ""
echo "------------------------------------------Raw Dog Run------------------------------------------"
echo ""
echo "Run 'sudo go run server/server.go' to start the C2 server"
echo ""
echo "Run 'sudo go run api/api.go' to start the api sever"
echo ""
echo "Run 'cd ui && npm start' to start start the web client"
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
echo "See .env and ui/.env to tweak enviroment variables (not advised)"
echo ""

