package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/99xtal/frinkiac-bot/frinkiac"
	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session
var frinkiacClient *frinkiac.FrinkiacClient
var err error

var (
	applicationID = os.Getenv("APPLICATION_ID")
	botAPIToken = os.Getenv("DISCORD_API_TOKEN")
)

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

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	"frinkiac": func (s *discordgo.Session, i *discordgo.InteractionCreate) error {
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
		for _, option := range(i.ApplicationCommandData().Options) {
			optionMap[option.Name] = option
		}

		searchQuery := optionMap["query"]
		frames, err := frinkiacClient.Search(searchQuery.StringValue())
		if err != nil {
			return err
		}

		if len(frames) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No frames found for search query: '" + searchQuery.StringValue() +"'",
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			return nil
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Image: &discordgo.MessageEmbedImage{
							URL: frames[0].GetPhotoUrl(),
						},
					},
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

	frinkiacClient = frinkiac.NewFrinkiacClient()

	registerCommands()
}

func main() {
	defer s.Close()
	s.AddHandler(func (s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is running")
	})

	s.AddHandler(func (s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			err = commandHandlers[i.ApplicationCommandData().Name](s, i)
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