#!/bin/bash

set -euo pipefail
trap 'echo "Error occurred on line $LINENO. Exiting."; exit 1' ERR

# Check for root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root."
    exit 1
fi

# Default values
build_type="local"
tag="snapshot"
build_type="local"
tag="snapshot"
PORT="8081"
WEBSERVER_PORT="8000"
REACT_APP_NGINX_PORT="8443"
C2SERVER_PORT="9000"
REDIRECTOR_PORT="9000"
DB_PORT="5432"
DB_USER="patron"

function show_help {
   echo "Usage: $0 [options]"
   echo "Options:"
   echo "  -l                 Use local build (default)"
   echo "  -r                 Use release build"
   echo "  -t <tag>           Set image tag (required for release)"
   echo "  --port <port>                     (default: 8081)"
   echo "  --web-port <port>                (default: 8000)"
   echo "  --nginx-port <port>              (default: 8443)"
   echo "  --c2-port <port>                 (default: 9000)"
   echo "  --redirector-port <port>         (default: 9000)"
   echo "  --db-port <port>                 (default: 5432)"
   echo "  --db-user <user>                 (default: patron)"
   echo "  -h                 Show this help message"
}

# Use getopt for long options
TEMP=$(getopt -o lrt:h --long port:,web-port:,nginx-port:,c2-port:,redirector-port:,db-port:,db-user: -n "$0" -- "$@")
eval set -- "$TEMP"

while true; do
   case "$1" in
      -l ) build_type="local"; shift ;;
      -r ) build_type="release"; shift ;;
      -t ) tag="$2"; shift 2 ;;
      --port ) PORT="$2"; shift 2 ;;
      --web-port ) WEBSERVER_PORT="$2"; shift 2 ;;
      --nginx-port ) REACT_APP_NGINX_PORT="$2"; shift 2 ;;
      --c2-port ) C2SERVER_PORT="$2"; shift 2 ;;
      --redirector-port ) REDIRECTOR_PORT="$2"; shift 2 ;;
      --db-port ) DB_PORT="$2"; shift 2 ;;
      --db-user ) DB_USER="$2"; shift 2 ;;
      -h ) show_help; exit 0 ;;
      -- ) shift; break ;;
      * ) break ;;
   esac
done

# Prompt user for build type and tag
echo "Do you want to perform a local or release build?"
select build_type in "local" "release"; do
    if [[ "$build_type" == "local" || "$build_type" == "release" ]]; then
        break
    else
        echo "Invalid selection."
    fi
done

# Prompt for tag override
if [[ "$build_type" == "local" ]]; then
    read -p "Enter a tag to use [default: snapshot]: " user_tag
    tag=${user_tag:-snapshot}
else
    while [[ "$tag" == "snapshot" ]]; do
        read -p "Enter a version tag for release (e.g. v1.0.0): " tag
        if [[ -z "$tag" || "$tag" == "snapshot" ]]; then
            echo "❌ Release builds must not use 'snapshot'"
        fi
    done
fi

# Prerequisite check
function prereq_check {
    for cmd in base64 openssl docker make; do
        if ! command -v $cmd &>/dev/null; then
            echo "Error: $cmd is not installed."
            exit 1
        fi
    done
    if ! docker info >/dev/null 2>&1; then
        echo "Error: Docker is not running."
        exit 1
    fi
}

# Set proxy variables
function set_proxy_variables {
    read -p "Enter HTTP Proxy (leave blank if none): " http_proxy
    read -p "Enter HTTPS Proxy (leave blank if none): " https_proxy
    read -p "Enter NO Proxy (e.g. localhost,127.0.0.1): " no_proxy

    export HTTP_PROXY="${http_proxy:-}"
    export HTTPS_PROXY="${https_proxy:-}"
    export NO_PROXY="${no_proxy:-}"

    echo -e "\nUsing proxy settings:"
    echo "HTTP_PROXY=$HTTP_PROXY"
    echo "HTTPS_PROXY=$HTTPS_PROXY"
    echo "NO_PROXY=$NO_PROXY"
}

# Run prechecks and proxy config
prereq_check
set_proxy_variables

# Write environment
echo "📝 Writing .env for Docker Bake..."
cat <<EOF > .env
TAG=$tag
PORT=$PORT
WEBSERVER_PORT=$WEBSERVER_PORT
REACT_APP_NGINX_PORT=$REACT_APP_NGINX_PORT
C2SERVER_PORT=$C2SERVER_PORT
REDIRECTOR_PORT=$REDIRECTOR_PORT
DB_PORT=$DB_PORT
DB_USER=$DB_USER
HTTP_PROXY=$HTTP_PROXY
HTTPS_PROXY=$HTTPS_PROXY
NO_PROXY=$NO_PROXY
EOF

export $(grep -v '^#' .env | xargs)

echo -e "\n🚀 Running docker buildx bake for: $build_type"
docker buildx bake $build_type
