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
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if MessageContains(m.Content, "but") {
		if ShouldSendMessage() {
			s.ChannelMessageSend(m.ChannelID, "butt")
		}
	}
}

func come(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if MessageContains(m.Content, "come") || MessageContains(m.Content, "coming") || MessageContains(m.Content, "cum") {
		if ShouldSendMessage() {
			s.ChannelMessageSend(m.ChannelID, "no coming")
		}
	}
}

func cow(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if MessageContains(m.Content, "cow") {
		s.ChannelMessageSend(m.ChannelID, "MOOOOO! "+"<:rage:1302882593339084851>")
	}
}

func dog(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if MessageContains(m.Content, "dog") {
		s.ChannelMessageSend(m.ChannelID, "what the dog doin'?")
	}
}

func egg(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if MessageContains(m.Content, "egg") {
		s.ChannelMessageSend(m.ChannelID, "egg")
	}
}

func newMember(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Type == discordgo.MessageTypeGuildMemberJoin {
		s.MessageReactionAdd(m.ChannelID, m.ID, "<a:catJAM:1111234741639848026>")
		s.MessageReactionAdd(m.ChannelID, m.ID, "<a:hooray:1057490323561001042>")
		s.MessageReactionAdd(m.ChannelID, m.ID, "<a:pedro:1057490323561001042>")
	}
}

func ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	if ShouldIgnoreMessage(s, m) {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

func GuildMemberRemove(s *discordgo.Session, gm *discordgo.GuildMemberRemove) {
	s.ChannelMessageSend("1018070065423335437", gm.User.Username+" has left the server. <:periodt:1302882591552307240>")
}
