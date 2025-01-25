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

target "_common" {
    args {
        HTTP_PROXY   = "${HTTP_PROXY}"
        HTTPS_PROXY  = "${HTTPS_PROXY}"
        NO_PROXY     = "${NO_PROXY}"
    }
    context = "."
}

target "nginx-local" {
    dockerfile = "Dockerfile.nginx"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
    }
}

target "postgres-local" {
    dockerfile = "Dockerfile.postgres"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
    }
}

target "server-local" {
    dockerfile = "Dockerfile.server"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
    }
}

target "api-local" {
    dockerfile = "Dockerfile.api"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
    }
}

target "ui-local" {
    dockerfile = "Dockerfile.ui"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
    }
}

target "redirector-local" {
    dockerfile = "Dockerfile.redirector"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${REDIRECTOR}:${TAG}"]
    args = {
      REDIRECTOR_PORT = "${REDIRECTOR_PORT}"
    }
}

target "bot-local" {
    dockerfile = "./bot/Dockerfile.bot"
    inherits = [ "_common" ]
    output = ["type=docker"]
    tags = ["${BOT}:${TAG}"]
    args = {
      REPO_URL = "https://github.com/PatronC2/PatronCLI.git"
      REPO_BRANCH = "main"
    }
}

target "nginx-release" {
    dockerfile = "Dockerfile.nginx"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${NGINX}:${TAG}"]
    tags = ["${REGISTRY}/${NGINX}:${TAG}"]
    args = {
      REACT_APP_NGINX_PORT = "${REACT_APP_NGINX_PORT}"
    }
}

target "postgres-release" {
    dockerfile = "Dockerfile.postgres"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${POSTGRES}:${TAG}"]
    tags = ["${REGISTRY}/${POSTGRES}:${TAG}"]
    args = {
      DB_PORT = "${DB_PORT}"
      DB_USER = "${DB_USER}"
    }
}

target "server-release" {
    dockerfile = "Dockerfile.server"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${SERVER}:${TAG}"]
    tags = ["${REGISTRY}/${SERVER}:${TAG}"]
    args = {
      C2SERVER_PORT = "${C2SERVER_PORT}"
    }
}

target "api-release" {
    dockerfile = "Dockerfile.api"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${API}:${TAG}"]
    tags = ["${REGISTRY}/${API}:${TAG}"]
    args = {
      WEBSERVER_PORT = "${WEBSERVER_PORT}"
    }
}

target "ui-release" {
    dockerfile = "Dockerfile.ui"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${UI}:${TAG}"]
    tags = ["${REGISTRY}/${UI}:${TAG}"]
    args = {
      PORT = "${PORT}"
    }
}

target "bot-release" {
    dockerfile = "./bot/Dockerfile.bot"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${BOT}:${TAG}"]
    tags = ["${REGISTRY}/${BOT}:${TAG}"]
    args = {
      REPO_URL = "https://github.com/PatronC2/PatronCLI.git"
      REPO_BRANCH = "main"
    }
}

target "redirector-release" {
    dockerfile = "Dockerfile.redirector"
    inherits = [ "_common" ]
    output = ["type=registry,output=registry.${REGISTRY}/${REDIRECTOR}:${TAG}"]
    tags = ["${REGISTRY}/${REDIRECTOR}:${TAG}"]
    args = {
      REDIRECTOR_PORT = "${REDIRECTOR_PORT}"
    }
}

group "local" {
    targets = ["nginx-local", "postgres-local", "api-local", "ui-local", "server-local", "redirector-local", "bot-local"]
}

group "default" {
    targets = ["nginx-release", "postgres-local", "api-release", "ui-release", "server-release", "redirector-release", "bot-release"]
}
