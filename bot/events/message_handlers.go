package events

import (
	"github.com/bwmarrin/discordgo"
)

var MessageHandlers = []func(*discordgo.Session, *discordgo.MessageCreate){
	butt,
	come,
	cow,
	dog,
	egg,
	newMember,
	ping,
}

func butt(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "but") {
		if chanceSucceeded(0.2) {
			s.ChannelMessageSend(m.ChannelID, "butt")
		}
	}
}

func come(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "come") || messageContains(m.Content, "coming") || messageContains(m.Content, "cum") {
		if chanceSucceeded(0.2) {
			s.ChannelMessageSend(m.ChannelID, "no coming")
		}
	}
}

func cow(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "cow") {
		s.ChannelMessageSend(m.ChannelID, "MOOOOO! "+"<:rage:1302882593339084851>")
	}
}

func dog(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "dog") {
		s.ChannelMessageSend(m.ChannelID, "what the dog doin'?")
	}
}

func egg(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "egg") {
		s.ChannelMessageSend(m.ChannelID, "egg")
	}
}

func newMember(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Type == discordgo.MessageTypeGuildMemberJoin {
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:catJAM:1111234741639848026")
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:hooray:1057490323561001042")
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:pedro:1057490323561001042")
	}
}

func ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}
