package components

import "github.com/bwmarrin/discordgo"

func ImageLinkEmbed(url string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: url,
		},
	}
}

func PreviewActionsComponent(startReached bool, endReached bool) discordgo.MessageComponent {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Style:    discordgo.SecondaryButton,
				Label:    "Previous",
				CustomID: "previous_result",
				Disabled: startReached,
			},
			discordgo.Button{
				Style:    discordgo.SecondaryButton,
				Label:    "Next",
				CustomID: "next_result",
				Disabled: endReached,
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
