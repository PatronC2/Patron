package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PatronC2/Patron/types"
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
	return sendMsg
}

func GetBotAgents() *discordgo.MessageSend {
	apiserver := goDotEnvVariable("WEBSERVER_IP")
	apiport := goDotEnvVariable("WEBSERVER_PORT")
	apiURL := fmt.Sprintf("http://%s:%s/api/agents", apiserver, apiport)
	// fmt.Println(apiURL)
	// Create new HTTP client & set timeout
	client := http.Client{Timeout: 5 * time.Second}

	// Query C2 API
	response, err := client.Get(apiURL)
	if err != nil {
		return &discordgo.MessageSend{
			Content: "Http api Error! trying to get agents",
		}
	}

	// Open HTTP response body
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	// Convert JSON
	var data types.BotConfigAgent
	json.Unmarshal([]byte(body), &data)
	// fmt.Println(data[0].Uuid)
	var results strings.Builder
	for i := range data {
		fmt.Fprintf(&results, "%s %s@%s %s\n", data[i].Uuid, data[i].Username, data[i].AgentIP, data[i].Hostname)
	}
	// fmt.Println(results.String())

	return sendAgentMessage(results.String()[:len([]rune(results.String()))], "Agent Info")
}

func GetBotAgent(message string) *discordgo.MessageSend {
	apiserver := goDotEnvVariable("WEBSERVER_IP")
	apiport := goDotEnvVariable("WEBSERVER_PORT")
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
	}

	apiURL := fmt.Sprintf("http://%s:%s/api/agent/%s", apiserver, apiport, botmsg[1])
	// fmt.Println(apiURL)
	// Create new HTTP client & set timeout
	client := http.Client{Timeout: 5 * time.Second}

	// Query C2 API
	response, err := client.Get(apiURL)
	if err != nil {
		return &discordgo.MessageSend{
			Content: "Http api Error! trying to get agent",
		}
	}

	// Open HTTP response body
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	// Convert JSON
	var data types.BotAgent
	json.Unmarshal([]byte(body), &data)
	// fmt.Println(data[0].Uuid)
	var results strings.Builder
	for i := range data {
		fmt.Fprintf(&results, "`Command:` %s \n`Output:` %s", data[i].Command, data[i].Output)
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
		return &discordgo.MessageSend{
			Content: "Charater limit reached (> 4096)",
		}
	}
}
