![GitHub last commit](https://img.shields.io/github/last-commit/PatronC2/Patron?style=flat&logo=github)
![Lines of code](https://img.shields.io/tokei/lines/github/PatronC2/Patron?style=flat&logo=github)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/PatronC2/Patron?style=flat&logo=github)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/PatronC2/Patron?style=flat&logo=go)

# Patron

A Command and Control Framework made in Go.

## Create Cert

* openssl ecparam -genkey -name prime256v1 -out certs/server.key
* openssl req -new -x509 -key server.key -out certs/server.pem -days 3650
* base64 -w 0 certs/server.key 

# Build server manually

* CGO_ENABLED=0 go build -o build/server server/server.go  
* CGO_ENABLED=0 go build -o build/webserver Web/server/webserver.go

* sudo CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.ServerIP=10.10.10.113 -X main.ServerPort=6969 -X main.CallbackFrequency=10 -X main.CallbackJitter=10" -o test client/client.go