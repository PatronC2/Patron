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
    dockerfile = "nginx/Dockerfile"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${NGINX}:${TAG}"]
    tags = ["${REGISTRY}/${NGINX}:${TAG}"]
}

target "server" {
    dockerfile = "server/Dockerfile"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
}

target "api" {
    dockerfile = "api/Dockerfile"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
}

target "ui" {
    dockerfile = "ui/Dockerfile"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
}

group "default" {
    targets = ["nginx", "api", "server"]
}