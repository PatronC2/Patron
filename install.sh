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

function set_proxy_variables {
   read -p "Enter HTTP Proxy (or leave blank if not using a proxy): " http_proxy
   read -p "Enter HTTPS Proxy (or leave blank if not using a proxy): " https_proxy
   read -p "Enter NO Proxy (e.g., localhost,127.0.0.1): " no_proxy

   export HTTP_PROXY=${http_proxy:-""}
   export HTTPS_PROXY=${https_proxy:-""}
   export NO_PROXY=${no_proxy:-""}

   echo "Using Proxy Settings:"
   echo "  HTTP_PROXY=$HTTP_PROXY"
   echo "  HTTPS_PROXY=$HTTPS_PROXY"
   echo "  NO_PROXY=$NO_PROXY"
}

function configure_docker_proxy {
    if [[ -n "$HTTP_PROXY" || -n "$HTTPS_PROXY" ]]; then
        echo "🔧 Configuring Docker systemd proxy..."

        mkdir -p /etc/systemd/system/docker.service.d

        cat <<EOF > /etc/systemd/system/docker.service.d/http-proxy.conf
[Service]
Environment="HTTP_PROXY=$HTTP_PROXY"
Environment="HTTPS_PROXY=$HTTPS_PROXY"
Environment="NO_PROXY=$NO_PROXY"
EOF

        echo "🔄 Reloading systemd and restarting Docker..."
        systemctl daemon-reexec
        systemctl daemon-reload
        systemctl restart docker

        echo "✅ Docker proxy settings applied."
    else
        echo "ℹ️ No proxy settings provided — skipping Docker systemd proxy config."
    fi
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
         echo "Attempting to install Docker on Ubuntu..."
         HTTP_PROXY="$HTTP_PROXY" HTTPS_PROXY="$HTTPS_PROXY" http_proxy="$HTTP_PROXY" https_proxy="$HTTPS_PROXY" NO_PROXY="$NO_PROXY" \
         ./install-docker-ubuntu.sh || {
            echo "Failed to install Docker on Ubuntu.";
            exit 1;
         }
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

set_proxy_variables

prereq_app_check

configure_docker_proxy

function generate_certificates {
  echo "Generating TLS certificate for Patron..."

  # Prompt for certificate details with defaults
  read -p "Country (C) [US]: " cert_country
  cert_country=${cert_country:-US}

  read -p "State (ST) [Maryland]: " cert_state
  cert_state=${cert_state:-Maryland}

  read -p "City (L) [Towson]: " cert_city
  cert_city=${cert_city:-Towson}

  read -p "Organization (O) [PatronC2]: " cert_org
  cert_org=${cert_org:-PatronC2}

  read -p "Organizational Unit (OU) [OffensiveOps]: " cert_ou
  cert_ou=${cert_ou:-OffensiveOps}

  read -p "Common Name (CN, e.g. domain.com) [patronc2.net]: " cert_cn
  cert_cn=${cert_cn:-patronc2.net}

  echo "Creating certs..."
  mkdir -p certs
  rm -f certs/server.key certs/server.pem

  openssl ecparam -genkey -name prime256v1 -out certs/server.key

  openssl req -new -x509 \
    -key certs/server.key \
    -out certs/server.pem \
    -days 3650 \
    -subj "/C=$cert_country/ST=$cert_state/L=$cert_city/O=$cert_org/OU=$cert_ou/CN=$cert_cn"

  echo "✅ Certificate generated and saved to certs/server.pem"
}

generate_certificates

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
HTTP_PROXY=$http_proxy
HTTPS_PROXY=$https_proxy
NO_PROXY=$no_proxy
EOF

export $(grep -v '^#' .env | xargs)

echo "Generating frontend runtime config (frontend-config.json)..."

cat <<EOF > frontend-config.json
{
  "REACT_APP_C2SERVER_PORT": "$REACT_APP_C2SERVER_PORT",
  "REACT_APP_NGINX_PORT": "$REACT_APP_NGINX_PORT",
  "REACT_APP_NGINX_IP": "$REACT_APP_NGINX_IP",
  "REACT_SERVER_IP": "$REACT_SERVER_IP"
}
EOF

echo "✅ Created frontend-config.json"

echo "Starting Patron C2"
docker compose up -d

echo "------------------------------------------ Informational --------------------------------------"
echo ""
echo "✅ Patron C2 Install successful"
echo ""
echo "Visit https://$nginxip:$nginxport for Web"
echo ""
echo "C2 Server on $nginxip:$c2serverport"
echo ""
echo "See .env to tweak environment variables (not advised and restart required)"
echo "Run 'docker compose down --rmi all -v --remove-orphans' to stop"
echo "Run 'docker compose up -d' to restart"
echo ""
