package components

import (
	"github.com/99xtal/frinkiac-bot/pkg/session"
	"github.com/bwmarrin/discordgo"
)

func ImageLinkEmbed(url string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: url,
		},
	}
}

func PreviewActionsComponent(s *session.FrinkiacSession) discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Style:    discordgo.SecondaryButton,
				Label:    "Previous",
				CustomID: "previous_result",
				Disabled: s.Cursor == 0,
			},
			discordgo.Button{
				Style:    discordgo.SecondaryButton,
				Label:    "Next",
				CustomID: "next_result",
				Disabled: s.Cursor == len(s.SearchResults),
			},
			discordgo.Button{
				Style:    discordgo.SuccessButton,
				Label:    "Make Meme",
				CustomID: "open_meme_modal",
			},
			discordgo.Button{
				Style:    discordgo.PrimaryButton,
				Label:    "Send",
				CustomID: "send_frame",
			},
		},
	}
}
