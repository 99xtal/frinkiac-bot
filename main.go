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

type InteractionSession struct {
	cursor	int;
	frames []frinkiac.Frame
}

func (s *InteractionSession) cursorNext() {
	s.cursor += 1;
}

func (s *InteractionSession) cursorPrev() {
	if s.cursor > 0 {
		s.cursor -= 1;
	}
}

var interactionSessions map[string]*InteractionSession;

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

func createFrinkiacSession(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, option := range(i.ApplicationCommandData().Options) {
		optionMap[option.Name] = option
	}

	frameIndex := 0;
	searchQuery := optionMap["query"]
	frames, err := frinkiacClient.Search(searchQuery.StringValue())
	if err != nil {
		return err
	}

	log.Println(i.ID)
	interactionSessions[i.ID] = &InteractionSession{
		cursor: 0,
		frames: frames,
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
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: frames[frameIndex].GetPhotoUrl(),
					},
				},
			},
			Content: frames[frameIndex].Episode,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Previous",
							CustomID: "previous_result",
							Disabled: frameIndex == 0,
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Next",
							CustomID: "next_result",
							Disabled: frameIndex == len(frames) - 1,
						},
						discordgo.Button{
							Style: discordgo.PrimaryButton,
							Label: "Send",
							CustomID: "send_frame",
						},
					},
				},
			},
		},
	})

	return nil	
}

var applicationCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	"frinkiac": createFrinkiacSession,
}

func init() {
	authorization := fmt.Sprintf("Bot %s", botAPIToken)
	s, err = discordgo.New(authorization)
	if err != nil {
		log.Fatal(err)
	}

	frinkiacClient = frinkiac.NewFrinkiacClient()
	interactionSessions = make(map[string]*InteractionSession)

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
			interactionSessionData := interactionSessions[i.Message.Interaction.ID]

			switch i.MessageComponentData().CustomID {
			case "next_result":
				interactionSessions[i.Message.Interaction.ID].cursorNext()	
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{
							{
								Image: &discordgo.MessageEmbedImage{
									URL: interactionSessionData.frames[interactionSessionData.cursor].GetPhotoUrl(),
								},
							},
						},
						Content: interactionSessionData.frames[interactionSessionData.cursor].Episode,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Style: discordgo.SecondaryButton,
										Label: "Previous",
										CustomID: "previous_result",
										Disabled: interactionSessionData.cursor == 0,
									},
									discordgo.Button{
										Style: discordgo.SecondaryButton,
										Label: "Next",
										CustomID: "next_result",
										Disabled: interactionSessionData.cursor == len(interactionSessionData.frames) - 1,
									},
									discordgo.Button{
										Style: discordgo.PrimaryButton,
										Label: "Send",
										CustomID: "send_frame",
									},
								},
							},
						},
					},
				})
			case "previous_result":
				interactionSessionData.cursorPrev()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{
							{
								Image: &discordgo.MessageEmbedImage{
									URL: interactionSessionData.frames[interactionSessionData.cursor].GetPhotoUrl(),
								},
							},
						},
						Content: interactionSessionData.frames[interactionSessionData.cursor].Episode,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Style: discordgo.SecondaryButton,
										Label: "Previous",
										CustomID: "previous_result",
										Disabled: interactionSessionData.cursor == 0,
									},
									discordgo.Button{
										Style: discordgo.SecondaryButton,
										Label: "Next",
										CustomID: "next_result",
										Disabled: interactionSessionData.cursor == len(interactionSessionData.frames) - 1,
									},
									discordgo.Button{
										Style: discordgo.PrimaryButton,
										Label: "Send",
										CustomID: "send_frame",
									},
								},
							},
						},
					},
				})
			case "send_frame":
				s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Image: &discordgo.MessageEmbedImage{
									URL: interactionSessionData.frames[interactionSessionData.cursor].GetPhotoUrl(),
								},
							},
						},
					},
				})	
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