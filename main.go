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
		searchQuery := i.ApplicationCommandData().Options[0]
		session, err := NewFrinkiacSession(searchQuery.StringValue(), s)
		if err != nil {
			return err
		}
		interactionSessions[i.ID] = session
	
		if len(session.searchResults) == 0 {
			session.RespondWithEphemeralError(i.Interaction, "No frames found for search query: '" + searchQuery.StringValue() +"'")
			return nil
		}
	
		session.CreateMessagePreview(i.Interaction)
		return nil	
	},
}

var messageComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)  error {
	"next_result": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		messageSession.NextPage()
		messageSession.UpdateMessagePreview(i.Interaction)
		return nil
	},
	"previous_result": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		messageSession.PrevPage()
		messageSession.UpdateMessagePreview(i.Interaction)
		return nil
	},
	"send_frame": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
		messageSession.SubmitMessage(i.Interaction)	
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
			err = messageComponentHandlers[i.MessageComponentData().CustomID](s, i)
			if err != nil {
				log.Printf("Error: %v", err)
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