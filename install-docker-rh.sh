#!/bin/bash

# Remove conflicting Docker packages
for pkg in docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate podman podman-docker containerd runc; do
  sudo yum remove -y $pkg
done

# Install Docker
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo systemctl enable --now docker
sudo docker --version
