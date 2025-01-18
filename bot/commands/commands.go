package commands

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

var manageMessagesPermission int64 = discordgo.PermissionManageMessages

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "bonk",
		Description: "Bonk someone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "The person to bonk",
				Required:    true,
			},
		},
	},
	{
		Name:        "burrito",
		Description: "Tuck someone into bed",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "The person to tuck in",
				Required:    true,
			},
		},
	},
	{
		Name:        "brain",
		Description: "Someone's brain not working?",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "The person whose brain is not working",
				Required:    true,
			},
		},
	},
	{
		Name:                     "clear",
		Description:              "Clear messages in a channel",
		DefaultMemberPermissions: &manageMessagesPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "The number of messages to clear",
				Required:    true,
				MinValue:    &[]float64{1}[0],
				MaxValue:    100,
			},
		},
	},
	{
		Name:        "coinflip",
		Description: "Flip a coin",
	},
	{
		Name:        "hug",
		Description: "Hug someone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "The person to hug",
				Required:    true,
			},
		},
	},
	{
		Name:        "smooch",
		Description: "Give someone a nice big smooch",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "The person to smooch",
				Required:    true,
			},
		},
	},
	{
		Name:        "tomato",
		Description: "Toss a tomato at someone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "name",
				Description: "Your target",
				Required:    true,
			},
		},
	},
}

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"bonk": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("<a:bonk:1310467659090886678> %s", user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"burrito": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("%s has tucked %s into a burrito blanket. awwww goodnight %s <:burritoblanket:1021275794678497291>", i.Member.User.GlobalName, user, user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"brain": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("Oh dear, it looks like %s's brain has stopped working... Please wait a moment while it restarts. <:rip:1057489640636035102>", user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		amount := int(i.ApplicationCommandData().Options[0].IntValue())
		messages, err := s.ChannelMessages(i.ChannelID, amount, "", "", "")
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error fetching messages",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		var messageIDs []string
		for _, m := range messages {
			messageIDs = append(messageIDs, m.ID)
		}

		err = s.ChannelMessagesBulkDelete(i.ChannelID, messageIDs)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error deleting messages",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Deleted %d messages", amount),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},

	"coinflip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		coin := []string{"Heads", "Tails"}
		response := coin[rand.Intn(len(coin))]

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"hug": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("_hugs %s_", user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"smooch": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("%s has given %s a big smooch. MWAHHH! <:cuddle:1299195758960054364>", i.Member.User.GlobalName, user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},

	"tomato": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.ApplicationCommandData().Options[0].UserValue(nil).Mention()
		response := fmt.Sprintf("%s threw a tomato at %s. tomato tomato tomato! <:rip:1057489640636035102>", i.Member.User.GlobalName, user)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},
}
