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
var sessionManager *session.SessionManager
var err error

var (
	applicationID = os.Getenv("APPLICATION_ID")
	botAPIToken   = os.Getenv("DISCORD_API_TOKEN")
)

var interactionSessions map[string]*session.FrinkiacSession

func registerCommands() error {
	data, err := os.ReadFile("commands.json")
	if err != nil {
		return err
	}

	var configCommands []discordgo.ApplicationCommand
	err = json.Unmarshal(data, &configCommands)
	if err != nil {
		return err
	}

	var commandPtrs []*discordgo.ApplicationCommand
	for _, cmd := range configCommands {
		commandPtrs = append(commandPtrs, &cmd)
	}

	_, err = s.ApplicationCommandBulkOverwrite(applicationID, "", commandPtrs)
	return err
}

var applicationCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
	"frinkiac": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		searchQuery := i.ApplicationCommandData().Options[0]
		searchResults, err := frinkiacClient.Search(searchQuery.StringValue())
		if err != nil {
			return err
		}
		frinkiacSession := session.NewFrinkiacSession()
		frinkiacSession.SearchResults = searchResults
		sessionManager.Set(i.ID, frinkiacSession)

		if len(frinkiacSession.SearchResults) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No frames found for search query: '" + searchQuery.StringValue() + "'",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return nil
		}

		caption, err := frinkiacClient.GetCaption(frinkiacSession.GetCurrentFrame().Episode, fmt.Sprint(frinkiacSession.GetCurrentFrame().Timestamp))
		if err != nil {
			return err
		}
		frinkiacSession.CurrentFrameCaption = &caption

		currentFrame := frinkiacSession.GetCurrentFrame()
		frinkiacSession.CurrentImageLink = currentFrame.GetPhotoUrl()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(frinkiacSession.CurrentImageLink),
				},
				Content: fmt.Sprintf("\"%s\"\nSeason %d / Episode %d", caption.Episode.Title, caption.Episode.Season, caption.Episode.EpisodeNumber),
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(frinkiacSession),
				},
			},
		})
		return nil
	},
}

var messageComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
	"next_result": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		interactionSession, err := sessionManager.Get(i.Message.Interaction.ID)
		if err != nil {
			return err
		}
		currentFrame, err := interactionSession.NextResult()
		if err != nil {
			fmt.Printf("%v", err)
		}
		caption, err := frinkiacClient.GetCaption(currentFrame.Episode, fmt.Sprint(currentFrame.Timestamp))
		if err != nil {
			return err
		}
		interactionSession.CurrentFrameCaption = &caption
		interactionSession.CurrentImageLink = currentFrame.GetPhotoUrl()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(interactionSession.CurrentImageLink),
				},
				Content: fmt.Sprintf("\"%s\"\nSeason %d / Episode %d", caption.Episode.Title, caption.Episode.Season, caption.Episode.EpisodeNumber),
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(interactionSession),
				},
			},
		})
		sessionManager.Set(i.Message.Interaction.ID, interactionSession)
		return nil
	},
	"previous_result": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		interactionSession, err := sessionManager.Get(i.Message.Interaction.ID)
		if err != nil {
			return err
		}
		currentFrame, err := interactionSession.PreviousResult()
		if err != nil {
			fmt.Printf("%v", err)
		}
		caption, err := frinkiacClient.GetCaption(currentFrame.Episode, fmt.Sprint(currentFrame.Timestamp))
		if err != nil {
			return err
		}
		interactionSession.CurrentFrameCaption = &caption
		interactionSession.CurrentImageLink = currentFrame.GetPhotoUrl()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(interactionSession.CurrentImageLink),
				},
				Content: fmt.Sprintf("\"%s\"\nSeason %d / Episode %d", caption.Episode.Title, caption.Episode.Season, caption.Episode.EpisodeNumber),
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(interactionSession),
				},
			},
		})
		sessionManager.Set(i.Message.Interaction.ID, interactionSession)
		return nil
	},
	"send_frame": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		interactionSession, err := sessionManager.Get(i.Message.Interaction.ID)
		if err != nil {
			return err
		}
		s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(interactionSession.CurrentImageLink),
				},
			},
		})
		sessionManager.Delete(i.Message.Interaction.ID)
		return nil
	},
	"open_meme_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		interactionSession, err := sessionManager.Get(i.Message.Interaction.ID)
		if err != nil {
			return err
		}
		currentFrameCaption := interactionSession.CurrentFrameCaption
		defaultCaption := ""
		for _, sub := range currentFrameCaption.Subtitles {
			defaultCaption += sub.Content + " "
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "generate_meme_modal",
				Title:    "Generate Meme",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: "caption",
								Label:    "Meme Text",
								Style:    discordgo.TextInputParagraph,
								Value:    defaultCaption,
								Required: true,
							},
						},
					},
				},
			},
		})
		return nil
	},
}

var modalSubmitHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
	"generate_meme_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		interactionSession, err := sessionManager.Get(i.Message.Interaction.ID)
		if err != nil {
			return err
		}
		currentFrame := interactionSession.GetCurrentFrame()
		modalComponents := i.ModalSubmitData().Components
		memeCaption := modalComponents[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		caption, err := frinkiacClient.GetCaption(currentFrame.Episode, fmt.Sprint(currentFrame.Timestamp))
		if err != nil {
			return err
		}
		interactionSession.CurrentImageLink = currentFrame.GetCaptionPhotoUrl(memeCaption)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					components.ImageLinkEmbed(interactionSession.CurrentImageLink),
				},
				Content: fmt.Sprintf("\"%s\"\nSeason %d / Episode %d", caption.Episode.Title, caption.Episode.Season, caption.Episode.EpisodeNumber),
				Components: []discordgo.MessageComponent{
					components.PreviewActionsComponent(interactionSession),
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
	sessionManager = session.NewSessionManager()

	registerCommands()
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is running")
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("InteractionCreate", "appID:"+i.AppID, "interactionId:"+i.ID, i.Type.String())
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
		case discordgo.InteractionModalSubmit:
			err = modalSubmitHandlers[i.ModalSubmitData().CustomID](s, i)
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
