#!/bin/bash

set -euo pipefail
trap 'echo "Error occurred on line $LINENO. Exiting."; exit 1' ERR

if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root"
   exit 1
fi

mkdir -p logs

function show_help {
   echo "Usage: $0 [-d ] [-w ] [ -s <your_ip_address> ] [-b]"
   echo "Options:"
   echo "  -d    Use default options"
   echo "  -w    Wipe Database"
   echo "  -s    <your_ip_address>   Server IP address"
   echo "  -p    Prompts you to enter passwords"
   echo "  -b    Prompt for Discord Bot Token and start the bot container"
   echo "  -h    Show this help message"
}

function set_global_default_variable {
   webserverport="8000"
   reactclientip="0.0.0.0"
   reactclientport="8081"
   c2serverip=""
   c2serverport="9000"
   redirectorport="9000"
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
   echo "Note: Webserver, ReactClient, DB, and C2 server can't be on the same port and must be an unused port"
   read -p "Enter APISERVER PORT: " webserverport
   read -p "Enter REACTCLIENT IP: " reactclientip
   read -p "Enter REACTCLIENT PORT: " reactclientport
   echo "Note: To listen on all interfaces, leave C2SERVER IP blank"
   read -p "Enter C2SERVER IP: " c2serverip
   read -p "Enter C2SERVER PORT: " c2serverport
   read -p "Enter DOCKER INTERNAL NETWORK e.g. 172.18.0 (without the last octet): " dockerinternal
   read -p "Enter NGINX PORT: " nginxport
   read -p "Enter Database Host: " dbhost
   read -p "Enter Database Port: " dbport
   read -p "Enter Database User: " dbuser
   read -p "Enter Database Name: " dbname
}

function prompt_bot_token {
   read -p "Enter your Discord Bot Token: " bottoken
   if [ -z "$bottoken" ]; then
      echo "Error: Discord Bot Token cannot be empty."
      exit 1
   fi
}

function run_bot_container {
   echo "Starting patron_c2_bot container..."
   export BOT_TOKEN=$bottoken
   echo "BOT_TOKEN=$bottoken" >> .env
   docker compose up -d patron_c2_bot
   echo "Discord bot container started successfully."
}

function wipe_db {
   rm -rf data/postgres_data
   echo "Database Wiped!"
   dbpass=$(openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12)
   patronPassword=$(openssl rand -base64 9 | tr -dc 'a-zA-Z0-9' | head -c 12)
}

function pass_prompt {
   read -p "Enter Database Password: " dbpass
   read -p "Enter UI Username: " patronUsername
   read -p "Enter UI Password: " patronPassword
}

function prereq_app_check {
   base64=$(which base64 || echo "not found")
   openssl=$(which openssl || echo "not found")
   docker=$(which docker || echo "not found")

   # Check base64
   if [ -x "$base64" ]; then
      echo "base64 Check OK"
   else
      echo "Error: base64 is not installed"
      exit 1
   fi

   # Check openssl
   if [ -x "$openssl" ]; then
      echo "openssl Check OK"
   else
      echo "Error: openssl is not installed"
      exit 1
   fi

   # Check docker
   if [ -x "$docker" ]; then
      echo "Docker Check OK"
   else
      echo "Docker is not installed. Checking if I can install it for you."
      if which apt-get &>/dev/null; then
         sudo ./install-docker-ubuntu.sh || { echo "Failed to install Docker on Ubuntu."; exit 1; }
      elif which yum &>/dev/null; then
         sudo ./install-docker-rh.sh || { echo "Failed to install Docker on Red Hat."; exit 1; }
      else
         echo "Error: Can't install Docker for you. Please install it manually."
         exit 1
      fi
   fi

   # Check if Docker is running
   if ! docker info > /dev/null 2>&1; then
      echo "Error: Docker daemon is not running. Please start Docker."
      exit 1
   else
      echo "Docker is running."
   fi
}

default=""
clean_db=""
ipaddress=""
postgres_pass=""
run_bot=""

# Parse command line arguments using getopts
while getopts "dws:pbh" opt; do
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
      b)
         prompt_bot_token
         run_bot="y"
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
   echo "Error: Set your IP with -s"
   show_help
   exit 1
else
   nginxip=$ipaddress
fi

if [ -z "$default" ]; then
   ask_prompt
fi

# Check for mutually exclusive -w and -p flags
if [[ -n "$postgres_pass" && -n "$clean_db" ]] || [[ -z "$postgres_pass" && -z "$clean_db" ]]; then
   echo "Error: Both -w and -p flags must be used separately, or at least one must be used."
   show_help
   exit 1
fi

# Shift the processed options
shift $((OPTIND-1))

prereq_app_check

# Generate certs
echo "Generating certs..."
[ ! -d "$PWD/certs" ] && mkdir certs
rm -rf certs/server.key certs/server.pem
openssl ecparam -genkey -name prime256v1 -out certs/server.key
openssl req -new -x509 -key certs/server.key -out certs/server.pem -days 3650 -subj "/C=US/ST=Maryland/L=Towson/O=Case Studies/OU=Offensive Op/CN=example.com"

# Set environment variables
echo "Setting environment variables..."
rm -rf .env ui/.env

encpubkey=$(base64 -w 0 certs/server.pem)
JWT_KEY=$(openssl rand -base64 32)
REPO_DIR=$(pwd)

cat <<EOF > .env
WEBSERVER_PORT=$webserverport
C2SERVER_IP=$c2serverip
C2SERVER_PORT=$c2serverport
PUBLIC_KEY=$encpubkey
DISCORD_BOT_TOKEN=$bottoken
DB_HOST=$dbhost
DB_PORT=$dbport
DB_USER=$dbuser
DB_PASS=$dbpass
DB_NAME=$dbname
DOCKER_INTERNAL=$dockerinternal
ADMIN_AUTH_USER=$patronUsername
ADMIN_AUTH_PASS=$patronPassword
JWT_KEY=$JWT_KEY
REPO_DIR=$REPO_DIR
REACT_APP_C2SERVER_PORT=$c2serverport
REACT_APP_NGINX_PORT=$nginxport
REACT_APP_NGINX_IP=$nginxip
REACT_SERVER_IP=$reactclientip
HOST=$reactclientip
PORT=$reactclientport
REDIRECTOR_PORT=$redirectorport
EOF

echo "Cooking the Steak..."
export $(grep -v '^#' .env | xargs)
docker buildx bake local
docker compose up -d

if [ "$run_bot" == "y" ]; then
   run_bot_container
fi

echo "------------------------------------------ Informational --------------------------------------"
echo ""
echo "Visit https://$nginxip:$nginxport for Web"
echo ""
echo "C2 Server on $nginxip:$c2serverport"
echo ""
echo "See .env to tweak environment variables (not advised and restart required)"
echo "Run 'docker compose down --rmi all -v --remove-orphans' to stop"
echo "Run 'docker compose up -d' to restart"
echo ""
