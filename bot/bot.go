package main

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/PatronC2/Patron/bot/command"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/s-christian/gollehs/lib/logger"
)

// add this to an helper function
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func newMsg(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// ignore bot self
	if message.Author.ID == discord.State.User.ID {
		return
	}
	switch {
	case strings.Contains(message.Content, "!help"):
		discord.ChannelMessageSend(message.ChannelID, "Use `!agents` to list agents\nUse `!refresh <uuid>` to get agent commands/refresh\nUse `!cmd <uuid> <command>` to issue commands to the agent\nUse `!keys <uuid>` to get keylogs")
	case strings.Contains(message.Content, "!refresh"):
		logger.Logf(logger.Info, "Bot received !refresh triggered :"+message.Content+"\n")
		agentBot := command.GetBotAgent(message.Content)
		discord.ChannelMessageSendComplex(message.ChannelID, agentBot)
	case strings.Contains(message.Content, "!agents"):
		logger.Logf(logger.Info, "Bot received !agents triggered :"+message.Content+"\n")
		agentsBot := command.GetBotAgents()
		discord.ChannelMessageSendComplex(message.ChannelID, agentsBot)
	case strings.Contains(message.Content, "milk"):
		discord.ChannelMessageSend(message.ChannelID, "I love milk")
	}
}

func main() {
	botToken := goDotEnvVariable("BOT_TOKEN")
	// create session
	logger.Logf(logger.Info, "Discord Bot Started\n")
	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		logger.Logf(logger.Debug, "Discord Bot Crashed\n")
	}

	// Add event handler
	discord.AddHandler(newMsg)

	discord.Open()
	defer discord.Close()

	logger.Logf(logger.Info, "Discord Running\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
