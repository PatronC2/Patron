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
	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func sendAgentMessage(agents string) *discordgo.MessageSend {
	sendMsg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Agent Info",
			Description: agents,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Agents",
					Value:  "Agents",
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
			Content: "Error! trying to get agents",
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
	fmt.Println(results.String())

	return sendAgentMessage(results.String()[:len([]rune(results.String()))])
}

// func GetBotAgents() *discordgo.MessageSend {
// 	apiserver := goDotEnvVariable("WEBSERVER_IP")
// 	apiport := goDotEnvVariable("WEBSERVER_PORT")
// 	apiURL := fmt.Sprintf("http://%s:%s/api/agents", apiserver, apiport)
// 	fmt.Println(apiURL)
// 	// Create new HTTP client & set timeout
// 	client := http.Client{Timeout: 5 * time.Second}

// 	// Query C2 API
// 	response, err := client.Get(apiURL)
// 	if err != nil {
// 		return &discordgo.MessageSend{
// 			Content: "Error! trying to get agents",
// 		}
// 	}

// 	// Open HTTP response body
// 	body, _ := ioutil.ReadAll(response.Body)
// 	defer response.Body.Close()

// 	// Convert JSON
// 	var data types.BotConfigAgent
// 	json.Unmarshal([]byte(body), &data)
// 	fmt.Println(data[0].Uuid)
// 	var results string
// 	for i := range data {
// 		results = fmt.Sprintf("%s@%s %s\n", data[i].Uuid, data[i].AgentIP, data[i].Hostname)
// 	}

// 	return &discordgo.MessageSend{
// 		Content: "Error! trying to get agents",
// 	}
// }

// func GetBotAgents() string {
// 	apiserver := goDotEnvVariable("WEBSERVER_IP")
// 	apiport := goDotEnvVariable("WEBSERVER_PORT")
// 	apiURL := fmt.Sprintf("http://%s:%s/api/agents", apiserver, apiport)
// 	// fmt.Println(apiURL)
// 	// Create new HTTP client & set timeout
// 	client := http.Client{Timeout: 5 * time.Second}

// 	// Query C2 API
// 	response, err := client.Get(apiURL)
// 	if err != nil {
// 		return "Error! trying to get agents"
// 	}

// 	// Open HTTP response body
// 	body, _ := ioutil.ReadAll(response.Body)
// 	defer response.Body.Close()

// 	// Convert JSON
// 	var data types.BotConfigAgent
// 	json.Unmarshal([]byte(body), &data)
// 	// fmt.Println(data[0].Uuid)
// 	var results strings.Builder
// 	for i := range data {
// 		fmt.Fprintf(&results, "%s %s@%s %s\n", data[i].Uuid, data[i].Username, data[i].AgentIP, data[i].Hostname)
// 	}
// 	fmt.Println(results.String())

// 	return results.String()[:2001]
// }
