#!/bin/bash

images_to_delete=(
    "patron_c2_server"
    "node:18-alpine"
    "postgres"
    "nginx"
    "golang:1.20-alpine"
)

for image in "${images_to_delete[@]}"; do
    if docker image inspect "$image" > /dev/null 2>&1; then
        containers=$(docker ps -a -q --filter ancestor="$image")
        if [ -n "$containers" ]; then
            echo "Stopping and removing containers for image: $image"
            docker stop $containers
            docker rm $containers
        else
            echo "No containers found for image: $image"
        fi

        echo "Deleting image: $image"
        docker image rm "$image"
    else
        echo "Image not found locally: $image"
    fi
done
