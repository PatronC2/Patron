variable "TAG" {
  default = "snapshot"
}

variable "REGISTRY" {
  default = "patronc2"
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

target "ui-base" {
  dockerfile = "Dockerfile.ui"
  context = "."
  args = {
    PORT = "${PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/ui:${TAG}",
    "${REGISTRY}/ui:latest"
  ]
}

target "nginx-base" {
  dockerfile = "Dockerfile.nginx"
  context = "."
  args = {
    REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
   NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/nginx:${TAG}",
    "${REGISTRY}/nginx:latest"
  ]
}

target "redirector-base" {
  dockerfile = "Dockerfile.redirector"
  context = "."
  args = {
    REDIRECTOR_PORT = "${REDIRECTOR_PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/redirector:${TAG}",
    "${REGISTRY}/redirector:latest"
  ]
}

target "server-base" {
  dockerfile = "Dockerfile.server"
  context = "."
  args = {
    C2SERVER_PORT = "${C2SERVER_PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/server:${TAG}",
    "${REGISTRY}/server:latest"
  ]
}

target "postgres-base" {
  dockerfile = "Dockerfile.postgres"
  context = "."
  args = {
    DB_PORT = "${DB_PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/postgres:${TAG}",
    "${REGISTRY}/postgres:latest"
  ]
}

target "api-base" {
  dockerfile = "Dockerfile.api"
  context = "."
  args = {
    WEBSERVER_PORT = "${WEBSERVER_PORT}"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/api:${TAG}",
    "${REGISTRY}/api:latest"
  ]
}

target "bot-base" {
  dockerfile = "./bot/Dockerfile.bot"
  context = "."
  args = {
    REPO_URL = "https://github.com/PatronC2/PatronCLI.git"
    REPO_BRANCH = "main"
    HTTP_PROXY = "${HTTP_PROXY}"
    HTTPS_PROXY = "${HTTPS_PROXY}"
    NO_PROXY = "${NO_PROXY}"
  }
  tags = [
    "${REGISTRY}/bot:${TAG}",
    "${REGISTRY}/bot:latest"
  ]
}

target "nginx-local" {
  inherits = ["nginx-base"]
  output = ["type=docker"]
}

target "postgres-local" {
  inherits = ["postgres-base"]
  output = ["type=docker"]
}

target "api-local" {
  inherits = ["api-base"]
  output = ["type=docker"]
}

target "ui-local" {
  inherits = ["ui-base"]
  output = ["type=docker"]
}

target "server-local" {
  inherits = ["server-base"]
  output = ["type=docker"]
}

target "redirector-local" {
  inherits = ["redirector-base"]
  output = ["type=docker"]
}

target "bot-local" {
  inherits = ["bot-base"]
  output = ["type=docker"]
}

target "nginx-release" {
  inherits = ["nginx-base"]
  output = ["type=registry"]
}

target "postgres-release" {
  inherits = ["postgres-base"]
  output = ["type=registry"]
}

target "api-release" {
  inherits = ["api-base"]
  output = ["type=registry"]
}

target "ui-release" {
  inherits = ["ui-base"]
  output = ["type=registry"]
}

target "server-release" {
  inherits = ["server-base"]
  output = ["type=registry"]
}

target "redirector-release" {
  inherits = ["redirector-base"]
  output = ["type=registry"]
}

target "bot-release" {
  inherits = ["bot-base"]
  output = ["type=registry"]
}

group "local" {
    targets = ["nginx-local", "postgres-local", "api-local", "ui-local", "server-local", "redirector-local", "bot-local"]
}

group "release" {
    targets = ["nginx-release", "postgres-release", "api-release", "ui-release", "server-release", "redirector-release", "bot-release"]
}

group "default" {
    targets = ["local"]
}
