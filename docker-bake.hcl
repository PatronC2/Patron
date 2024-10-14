variable "TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "hub.docker.com"
}

variable "NGINX" {
  default = "patron-nginx"
}

variable "SERVER" {
  default = "patron-server"
}

variable "API" {
  default = "patron-api"
}

variable "UI" {
  default = "patron-ui"
}

target "nginx" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${NGINX}:${TAG}"]
    tags = ["${REGISTRY}/${NGINX}:${TAG}"]
}

target "server" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
}

target "api" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
}

target "ui" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
}

group "default" {
    targets = ["nginx", "api", "ui", "server"]
}