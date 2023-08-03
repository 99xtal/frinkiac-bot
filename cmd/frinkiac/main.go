package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/99xtal/frinkiac-bot/pkg/api"
	"github.com/99xtal/frinkiac-bot/pkg/components"
	"github.com/99xtal/frinkiac-bot/pkg/session"
	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session
var frinkiacClient *api.FrinkiacClient
var err error

var (
	applicationID = os.Getenv("APPLICATION_ID")
	botAPIToken = os.Getenv("DISCORD_API_TOKEN")
)

var interactionSessions map[string]*session.FrinkiacSession;

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
		searchResults, err := frinkiacClient.Search(searchQuery.StringValue())
		if err != nil {
			return err
		}
		frinkiacSession := session.NewFrinkiacSession()
		frinkiacSession.SearchResults = searchResults
		interactionSessions[i.ID] = frinkiacSession
	
		if len(frinkiacSession.SearchResults) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No frames found for search query: '" + searchQuery.StringValue() + "'",
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			return nil
		}
	
		currentFrame := frinkiacSession.GetCurrentFrame() 
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(currentFrame.GetPhotoUrl()),
				},
				Content: currentFrame.Episode,
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(frinkiacSession.Cursor == 0, frinkiacSession.Cursor == len(frinkiacSession.SearchResults) - 1),
				},
			},
		})
		return nil	
	},
}

var messageComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)  error {
	"next_result": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		messageSession.NextPage()
		currentFrame := messageSession.GetCurrentFrame()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(currentFrame.GetPhotoUrl()),
				},
				Content: currentFrame.Episode,
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(messageSession.Cursor == 0, messageSession.Cursor == len(messageSession.SearchResults) - 1),
				},
			},
		})
		return nil
	},
	"previous_result": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		messageSession.PrevPage()
		currentFrame := messageSession.GetCurrentFrame()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(currentFrame.GetPhotoUrl()),
				},
				Content: currentFrame.Episode,
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(messageSession.Cursor == 0, messageSession.Cursor == len(messageSession.SearchResults) - 1),
				},
			},
		})
		return nil
	},
	"send_frame": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		messageSession := interactionSessions[i.Message.Interaction.ID]
		s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
		currentFrame := messageSession.GetCurrentFrame()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(currentFrame.GetPhotoUrl()),
				},
			},
		})
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
	interactionSessions = make(map[string]*session.FrinkiacSession)

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