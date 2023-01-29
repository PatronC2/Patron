![GitHub last commit](https://img.shields.io/github/last-commit/PatronC2/Patron?style=flat&logo=github)
![Lines of code](https://img.shields.io/tokei/lines/github/PatronC2/Patron?style=flat&logo=github)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/PatronC2/Patron?style=flat&logo=github)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/PatronC2/Patron?style=flat&logo=go)

# Patron

A Command and Control Framework made in Go.


# Features

* Functional web interface
* Keylogger
* TLS C2 communication
* Swappable/Flexible Agent
* Discord Bot


# Install

* Run `git clone https://github.com/PatronC2/Patron.git`
* Run `./install.sh`


## Install Notes


# Build server manually

* `CGO_ENABLED=0 sudo go build -o build/server server/server.go`  OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/server server/server.go `
* `CGO_ENABLED=0 sudo go build -o build/webserver Web/server/webserver.go` OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/webserver Web/server/webserver.go`
* * `CGO_ENABLED=0 go build -o build/webserver Web/server/webserver.go` OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/webserver bot/bot.go`

# Build agent manually 
* deprecated (needs publickey variable)
* sudo CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.ServerIP=10.10.10.113 -X main.ServerPort=6969 -X main.CallbackFrequency=10 -X main.CallbackJitter=10" -o test client/client.go
* sudo CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.ServerIP=10.10.10.113 -X main.ServerPort=6969 -X main.CallbackFrequency=10 -X main.CallbackJitter=10" -o test client/kclient/kclient.go

# Bugs


# Credits

* Web Template: Open Source Web Design (Insanity by dirac)
* * http://www.oswd.org/user/designs/id/22/
* Go Keylogger Library:  by MarinX
* * https://github.com/MarinX/keylogger
* Logging Utility: by Christian
* * https://github.com/s-christian/gollehs/lib/logger
