package commands

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

var (
	manageMessagesPermission int64 = discordgo.PermissionManageMessages
	administratorPermission  int64 = discordgo.PermissionAdministrator
)

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
		Name:                     "live",
		Description:              "Go live on Twitch",
		DefaultMemberPermissions: &administratorPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The message to display",
				Required:    true,
			},
		},
	},
	{
		Name:                     "poll",
		Description:              "Create a poll with up to 5 options",
		DefaultMemberPermissions: &manageMessagesPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "The poll question",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "1",
				Description: "Option 1",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "2",
				Description: "Option 2",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "3",
				Description: "Option 3",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "4",
				Description: "Option 4",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "5",
				Description: "Option 5",
				Required:    false,
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

	"live": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		message := i.ApplicationCommandData().Options[0].StringValue()
		response := fmt.Sprintf("%s @everyone https://twitch.tv/charsibel", message)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Going live, have fun!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error going live. Please try again.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_, err = s.ChannelMessageSend(i.ChannelID, response)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error sending live message. Please try again.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},

	"poll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		question := i.ApplicationCommandData().Options[0].StringValue()
		options := i.ApplicationCommandData().Options[1:]

		optionsList := ""
		for idx, option := range options {
			optionsList += fmt.Sprintf("\n%d. %s", idx+1, option.StringValue())
		}

		response := fmt.Sprintf("**%s**%s", question, optionsList)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Creating poll...",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error creating poll",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		message, err := s.ChannelMessageSend(i.ChannelID, response)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error sending poll message",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"}
		for idx := 0; idx < len(options); idx++ {
			err = s.MessageReactionAdd(message.ChannelID, message.ID, emojis[idx])
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error adding reactions to poll message",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
		}
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
