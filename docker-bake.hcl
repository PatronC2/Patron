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

variable "BOT" {
  default = "patron-bot"
}

variable "UI" {
  default = "patron-ui"
}

variable "REDIRECTOR" {
  default = "patron-redirector"
}

variable "WEBSERVER_PORT" {
  default = "8000"
}

variable "REDIRECTOR_PORT" {
  default = "9000"
}

variable "REACT_APP_NGINX_PORT" {
  default = "8443"
}

variable "DB_PORT" {
  default = "5432"
}

variable "DB_USER" {
  default = "patron"
}

variable "C2SERVER_PORT" {
  default = "9000"
}

variable "PORT" {
  default = "8081"
}

variable "HTTP_PROXY" {
  default = ""
}

variable "HTTPS_PROXY" {
  default = ""
}

variable "NO_PROXY" {
  default = ""
}

target "nginx-local" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=docker"]
    tags = ["${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "postgres-local" {
    dockerfile = "Dockerfile.postgres"
    context = "."
    output = ["type=docker"]
    tags = ["${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "server-local" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=docker"]
    tags = ["${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "api-local" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=docker"]
    tags = ["${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "ui-local" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=docker"]
    tags = ["${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "redirector-local" {
    dockerfile = "Dockerfile.redirector"
    context = "."
    output = ["type=docker"]
    tags = ["${REDIRECTOR}:${TAG}"]
    args = {
      REDIRECTOR_PORT = "${REDIRECTOR_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "bot-local" {
    dockerfile = "./bot/Dockerfile.bot"
    context = "."
    output = ["type=docker"]
    tags = ["${BOT}:${TAG}"]
    args = {
      REPO_URL = "https://github.com/PatronC2/PatronCLI.git"
      REPO_BRANCH = "main"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "nginx-release" {
    dockerfile = "Dockerfile.nginx"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${NGINX}:${TAG}"]
    tags = ["${REGISTRY}/${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "postgres-release" {
    dockerfile = "Dockerfile.postgres"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${POSTGRES}:${TAG}"]
    tags = ["${REGISTRY}/${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
      DB_USER = "${DB_USER}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "server-release" {
    dockerfile = "Dockerfile.server"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "api-release" {
    dockerfile = "Dockerfile.api"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "ui-release" {
    dockerfile = "Dockerfile.ui"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "bot-release" {
    dockerfile = "./bot/Dockerfile.bot"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${BOT}:${TAG}"]
    tags = ["${REGISTRY}/${BOT}:${TAG}"]
    args = {
      REPO_URL = "https://github.com/PatronC2/PatronCLI.git"
      REPO_BRANCH = "main"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

target "redirector-release" {
    dockerfile = "Dockerfile.redirector"
    context = "."
    output = ["type=registry,output=registry.${REGISTRY}/${REDIRECTOR}:${TAG}"]
    tags = ["${REGISTRY}/${REDIRECTOR}:${TAG}"]
    args = {
      REDIRECTOR_PORT = "${REDIRECTOR_PORT}"
      HTTP_PROXY = "${HTTP_PROXY}"
      HTTPS_PROXY = "${HTTPS_PROXY}"
      NO_PROXY = "${NO_PROXY}"
    }
}

group "local" {
    targets = ["nginx-local", "postgres-local", "api-local", "ui-local", "server-local", "redirector-local", "bot-local"]
}

group "default" {
    targets = ["nginx-release", "postgres-local", "api-release", "ui-release", "server-release", "redirector-release", "bot-release"]
}
