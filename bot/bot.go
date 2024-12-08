package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Credential struct {
	Profile string `json:"profile"`
	IP      string `json:"ip"`
	Port    string `json:"port"`
	Token   string `json:"token"`
}

const credentialsPath = ".patron/credentials"

var (
	botToken            = os.Getenv("DISCORD_BOT_TOKEN")
	applicationCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "configure",
			Description: "Save a user token",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "token",
					Description: "The token to save",
					Required:    true,
				},
			},
		},
		{
			Name:        "patron",
			Description: "Execute a patron command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "The command to execute",
					Required:    true,
				},
			},
		},
	}
)

func main() {
	log.Println("Starting bot...")

	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is not set")
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Printf("Interaction received: %s", i.ApplicationCommandData().Name)
		handleInteraction(s, i)
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	}
	defer dg.Close()

	log.Println("Bot connected to Discord successfully.")

	for _, cmd := range applicationCommands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Fatalf("Cannot create '%v' command globally: %v", cmd.Name, err)
		}
		log.Printf("Command '%v' registered globally.", cmd.Name)
	}

	log.Println("Bot is running. Press CTRL+C to exit.")
	select {}
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		log.Println("Interaction received in a DM.")
	} else {
		log.Printf("Interaction received in guild: %s", i.GuildID)
	}

	switch i.ApplicationCommandData().Name {
	case "configure":
		log.Println("Handling 'configure' command.")
		handleSaveCommand(s, i)
	case "patron":
		log.Println("Handling 'patron' command.")
		handlePatronCommand(s, i)
	default:
		log.Printf("Unhandled command: %s", i.ApplicationCommandData().Name)
		sendResponse(s, i, "Unknown command. Use /help for a list of available commands.")
	}
}

func handleSaveCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := getUserID(i)
	if err != nil {
		sendResponse(s, i, "Failed to retrieve user information.")
		log.Printf("Error retrieving user ID: %v", err)
		return
	}

	options := i.ApplicationCommandData().Options
	var token string
	for _, opt := range options {
		if opt.Name == "token" {
			token = opt.StringValue()
		}
	}
	log.Printf("Received token for user %s: %s", userID, token)

	if token == "" {
		log.Println("Token is empty.")
		sendResponse(s, i, "Token is required.")
		return
	}

	err = saveCredential(Credential{
		Profile: userID,
		IP:      os.Getenv("PATRON_IP"),
		Port:    os.Getenv("PATRON_PORT"),
		Token:   token,
	})
	if err != nil {
		log.Printf("Error saving credential: %v", err)
		sendResponse(s, i, "Failed to save token: "+err.Error())
		return
	}

	log.Println("Token saved successfully.")
	sendResponse(s, i, "Token saved successfully!")
}

func handlePatronCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := getUserID(i)
	if err != nil {
		sendResponse(s, i, "Failed to retrieve user information.")
		log.Printf("Error retrieving user ID: %v", err)
		return
	}

	options := i.ApplicationCommandData().Options
	var command string

	for _, opt := range options {
		if opt.Name == "command" {
			command = opt.StringValue()
		}
	}
	log.Printf("Received command: %s", command)

	bannedStrings := []string{"||", "&&", ";", "`", "--profile"}
	if containsAnySubstring(command, bannedStrings) {
		sendResponse(s, i, "Nice try. Try harder.")
		log.Printf("Blocked potentially malicious command: %s", command)
		return
	}

	fullCommand := fmt.Sprintf("/usr/bin/patron %s --profile %s", command, userID)

	cmd := exec.Command("sh", "-c", fullCommand)
	log.Printf("Executing command: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing patron command: %v", err)
		sendResponse(s, i, fmt.Sprintf("Failed to execute command: %s\nOutput: %s", err.Error(), string(output)))
		return
	}

	log.Printf("Command executed successfully. Output: %s", string(output))
	sendResponse(s, i, fmt.Sprintf("Command executed successfully!\nOutput:\n%s", string(output)))
}

func sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("Failed to send interaction response: %v", err)
	}
	log.Printf("Response sent: %s", content)
}

func saveCredential(newCred Credential) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	credentialsFile := filepath.Join(homeDir, credentialsPath)
	log.Printf("Saving credential to file: %s", credentialsFile)

	var credentials []Credential

	if _, err := os.Stat(credentialsFile); err == nil {
		data, err := os.ReadFile(credentialsFile)
		if err != nil {
			return fmt.Errorf("failed to read credentials file: %w", err)
		}
		if err := json.Unmarshal(data, &credentials); err != nil {
			return fmt.Errorf("failed to parse credentials file: %w", err)
		}
		log.Println("Existing credentials loaded.")
	}

	updated := false
	for i, cred := range credentials {
		if cred.Profile == newCred.Profile {
			credentials[i] = newCred
			updated = true
			break
		}
	}

	if !updated {
		credentials = append(credentials, newCred)
	}

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(credentialsFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	err = os.WriteFile(credentialsFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	log.Println("Credential saved successfully.")
	return nil
}

func getUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID, nil
	}
	if i.User != nil {
		return i.User.ID, nil
	}
	return "", fmt.Errorf("failed to retrieve user information")
}

func containsAnySubstring(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}
