![GitHub last commit](https://img.shields.io/github/last-commit/PatronC2/Patron?style=flat&logo=github)
![Lines of code](https://img.shields.io/tokei/lines/github/PatronC2/Patron?style=flat&logo=github)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/PatronC2/Patron&logo=github)
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
  -b    Set up the discord bot
  -p    Prompts you to enter passwords
  -h    Show this help message
```

# Bugs

# Creating Discord bot
* Go to discord dev portal
* create new application (not actual bot)
* go to bot tab
* add bot (username , icon, uncheck public bot, check message content intent)
* Set Installation Contexts to User Install and Guild Install
* grab token (reset token, save for later)
* generate url to add to server, with permissions like `https://discord.com/oauth2/authorize?client_id=<your-client-id>&permissions=8&scope=bot+applications.commands&permissions=139586766912`
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
