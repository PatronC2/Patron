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

target "nginx-local" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=docker"]
    tags = ["${NGINX}:${TAG}"]
}

target "server-local" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=docker"]
    tags = ["${SERVER}:${TAG}"]
}

target "api-local" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=docker"]
    tags = ["${API}:${TAG}"]
}

target "ui-local" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=docker"]
    tags = ["${UI}:${TAG}"]
}

target "nginx-release" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=docker"]
    tags = ["${NGINX}:${TAG}"]
}

target "server-release" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
}

target "api-release" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
}

target "ui-release" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
}

group "local" {
    targets = ["nginx-local", "api-local", "ui-local", "server-local"]
}

group "default" {
    targets = ["nginx-release", "api-release", "ui-release", "server-release"]
}