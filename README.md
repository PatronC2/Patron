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
* Wrapped in docker

# Install

* Run `git clone https://github.com/PatronC2/Patron.git`
* Run `./install.sh -d -p -s <your-ip>` for fresh install

```
Options:
  -d    Use default options
  -w    Wipe Database
  -s    <your_ip_address>   Server Ip address
  -p    Prompts you to enter passwords
  -h    Show this help message
```


# Docker

* Run the install steps above
* Run `docker compose up --remove-orphans` 
* Tear down `docker compose down`
* debuging logs: 
* * `docker compose logs patron_c2_api -f`
* * `docker compose logs patron_c2_frontend -f`
* * `docker compose logs patron_c2_server -f`
* * `docker compose logs patron_c2_postgres -f`
* * `docker compose logs patron_c2_nginx -f`
* restarting
* * `docker compose restart patron_c2_api`
* * `docker compose restart patron_c2_frontend`
* * `docker compose restart patron_c2_server`
* * `docker compose restart patron_c2_postgres`
* * `docker compose restart patron_c2_nginx`

## Todo
* conditionally setup discord bot
* list / pull files with agents
* test multiple server 



## Install Notes


# Build server manually

* `CGO_ENABLED=0 sudo go build -o build/server server/server.go`  OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/server server/server.go `
* `CGO_ENABLED=0 sudo go build -o build/webserver Web/server/webserver.go` OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/webserver Web/server/webserver.go`
* * `CGO_ENABLED=0 go build -o build/webserver Web/server/webserver.go` OR `sudo go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o build/bot bot/bot.go`

# Build agent manually 
* deprecated (needs publickey variable)
* sudo CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.ServerIP=10.10.10.113 -X main.ServerPort=6969 -X main.CallbackFrequency=10 -X main.CallbackJitter=10" -o test client/client.go
* sudo CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.ServerIP=10.10.10.113 -X main.ServerPort=6969 -X main.CallbackFrequency=10 -X main.CallbackJitter=10" -o test client/kclient/kclient.go

# Bugs

# Creating Discord bot
* Go to discord dev portal
* create new application (not actual bot)
* go to bot tab
* add bot (username , icon, uncheck public bot, check message content intent)
* grab token (reset token, save for later)
* generate url to add to server
* * go to oauth2 -> url generator (check bot)
* * check (read messages/view Channels, send messages, embed links, send messages in thread) #VERY IMPORTANT
* * copy and open url in new tab
* * add bot to each server

# Credits

* Web Template: Open Source Web Design (Insanity by dirac)
* * http://www.oswd.org/user/designs/id/22/
* Go Keylogger Library:  by MarinX
* * https://github.com/MarinX/keylogger
* Logging Utility: by Christian
* * https://github.com/s-christian/gollehs/lib/logger
