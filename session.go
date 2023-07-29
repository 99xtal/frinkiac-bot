package main

import (
	"github.com/99xtal/frinkiac-bot/frinkiac"
	"github.com/bwmarrin/discordgo"
)

type FrinkiacSession struct {
	session	 *discordgo.Session
	client *frinkiac.FrinkiacClient
	cursor int
	searchResults []*frinkiac.Frame
}

func (s *FrinkiacSession) NextPage() {
	if (s.cursor < len(s.searchResults)) {
		s.cursor += 1
	}
}

func (s *FrinkiacSession) PrevPage() {
	if (s.cursor > 0) {
		s.cursor -= 1
	}
}

func (s *FrinkiacSession) RespondWithEphemeralError(i *discordgo.Interaction, errorText string) {
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: errorText,
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func (s *FrinkiacSession) RespondWithNewEditView(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: currentFrame.GetPhotoUrl(),
					},				
				},
			},
			Content: currentFrame.Episode,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Previous",
							CustomID: "previous_result",
							Disabled: s.cursor == 0,
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Next",
							CustomID: "next_result",
							Disabled: s.cursor == len(s.searchResults) - 1,
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
}

func (s *FrinkiacSession) UpdateEditView(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: currentFrame.GetPhotoUrl(),
					},
				},
			},
			Content: currentFrame.Episode,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Previous",
							CustomID: "previous_result",
							Disabled: s.cursor == 0,
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Label: "Next",
							CustomID: "next_result",
							Disabled: s.cursor == len(s.searchResults) - 1,
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
}

func (s *FrinkiacSession) SubmitFrame(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: currentFrame.GetPhotoUrl(),
					},
				},
			},
		},
	})
}

func NewFrinkiacSession(query string, s *discordgo.Session) (*FrinkiacSession, error) {
	frames, err := frinkiacClient.Search(query)
	if err != nil {
		return nil, err
	}
	return &FrinkiacSession{
		session: s,
		client: frinkiacClient,
		cursor: 0,
		searchResults: frames,
	}, nil
}