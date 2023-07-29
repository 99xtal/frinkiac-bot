package main

import (
	"github.com/99xtal/frinkiac-bot/api"
	"github.com/bwmarrin/discordgo"
)

type FrinkiacSession struct {
	session	 *discordgo.Session
	client *api.FrinkiacClient
	cursor int
	memeMode bool
	searchResults []*api.Frame
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

func (s *FrinkiacSession) toggleMemeMode() {
	s.memeMode = !s.memeMode;
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

func (s *FrinkiacSession) CreateMessagePreview(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				ImageLinkEmbed(currentFrame.GetPhotoUrl()),
			},
			Content: currentFrame.Episode,
			Components: []discordgo.MessageComponent{
				PreviewActionsComponent(s.cursor == 0, s.cursor == len(s.searchResults) - 1),
			},
		},
	})
}

func (s *FrinkiacSession) UpdateMessagePreview(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				ImageLinkEmbed(currentFrame.GetPhotoUrl()),
			},
			Content: currentFrame.Episode,
			Components: []discordgo.MessageComponent{
				PreviewActionsComponent(s.cursor == 0, s.cursor == len(s.searchResults) - 1),
			},
		},
	})
}

func (s *FrinkiacSession) SubmitMessage(i *discordgo.Interaction) {
	currentFrame := s.searchResults[s.cursor]
	s.session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				ImageLinkEmbed(currentFrame.GetPhotoUrl()),
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
		memeMode: false,
	}, nil
}