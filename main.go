package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/99xtal/frinkiac-bot/api"
	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session
var frinkiacClient *api.FrinkiacClient
var err error

var (
	applicationID = os.Getenv("APPLICATION_ID")
	botAPIToken = os.Getenv("DISCORD_API_TOKEN")
)

var interactionSessions map[string]*FrinkiacSession;

func registerCommands() error {
	data, err := os.ReadFile("commands.json")
	if err != nil {
		return err
	}

	var configCommands []discordgo.ApplicationCommand;
	err = json.Unmarshal(data, &configCommands)
	if err != nil {
		return err
	}

	var commandPtrs []*discordgo.ApplicationCommand
	for _, cmd := range(configCommands) {
		commandPtrs = append(commandPtrs, &cmd)
	}

	_, err = s.ApplicationCommandBulkOverwrite(applicationID, "", commandPtrs)
	return err
}

var applicationCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	"frinkiac": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
		for _, option := range(i.ApplicationCommandData().Options) {
			optionMap[option.Name] = option
		}
	
		searchQuery := optionMap["query"]
		session, err := NewFrinkiacSession(searchQuery.StringValue(), s)
		if err != nil {
			return err
		}
		interactionSessions[i.ID] = session
	
		if len(session.searchResults) == 0 {
			session.RespondWithEphemeralError(i.Interaction, "No frames found for search query: '" + searchQuery.StringValue() +"'")
			return nil
		}
	
		session.RespondWithNewEditView(i.Interaction)
		return nil	
	},
}

func init() {
	authorization := fmt.Sprintf("Bot %s", botAPIToken)
	s, err = discordgo.New(authorization)
	if err != nil {
		log.Fatal(err)
	}

	frinkiacClient = api.NewFrinkiacClient()
	interactionSessions = make(map[string]*FrinkiacSession)

	registerCommands()
}

func main() {
	s.AddHandler(func (s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is running")
	})

	s.AddHandler(func (s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("InteractionCreate", i.AppID, i.Type.String())
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			err = applicationCommandHandlers[i.ApplicationCommandData().Name](s, i)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		case discordgo.InteractionMessageComponent:
			messageSession := interactionSessions[i.Message.Interaction.ID]

			switch i.MessageComponentData().CustomID {
			case "next_result":
				messageSession.NextPage()
				messageSession.UpdateEditView(i.Interaction)
			case "previous_result":
				messageSession.PrevPage()
				messageSession.UpdateEditView(i.Interaction)
			case "send_frame":
				s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
				messageSession.SubmitFrame(i.Interaction)
			}
		}
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}