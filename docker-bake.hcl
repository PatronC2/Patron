variable "TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "hub.docker.com"
}

variable "NGINX" {
  default = "patron-nginx"
}

variable "POSTGRES" {
  default = "patron-postgres"
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

variable "WEBSERVER_PORT" {
  default = "8000"
}

variable "REACT_APP_NGINX_PORT" {
  default = "8443"
}

variable "DB_PORT" {
  default = "5432"
}

variable "C2SERVER_PORT" {
  default = "9000"
}

variable "PORT" {
  default = "8081"
}

target "nginx-local" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=docker"]
    tags = ["${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
    }
}

target "postgres-local" {
    dockerfile = "Dockerfile.postgres"
    context = "."
    output = ["type=docker"]
    tags = ["${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
    }
}

target "server-local" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=docker"]
    tags = ["${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
    }
}

target "api-local" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=docker"]
    tags = ["${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
    }
}

target "ui-local" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=docker"]
    tags = ["${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
    }
}

target "nginx-release" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${NGINX}:${TAG}"]
    tags = ["${REGISTRY}/${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
    }
}

target "postgres-release" {
    dockerfile = "Dockerfile.postgres"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${POSTGRES}:${TAG}"]
    tags = ["${REGISTRY}/${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
    }
}

target "server-release" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
    }
}

target "api-release" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
    }
}

target "ui-release" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
    }
}

group "local" {
    targets = ["nginx-local", "postgres-local", "api-local", "ui-local", "server-local"]
}

group "default" {
    targets = ["nginx-release", "postgres-local", "api-release", "ui-release", "server-release"]
}