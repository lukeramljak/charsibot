package events

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	r               = rand.New(rand.NewSource(time.Now().UnixNano()))
	chanceToSend    = 0.5
	MessageHandlers = []func(*discordgo.Session, *discordgo.MessageCreate){
		Butt,
		Come,
		Cow,
		Ping,
	}
)

func Butt(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldIgnoreMessage(s, m) {
		return
	}

	if messageContains(m.Content, "but") {
		if shouldSendMessage() {
			s.ChannelMessageSend(m.ChannelID, "butt")
		}
	}
}

func Come(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldIgnoreMessage(s, m) {
		return
	}

	if messageContains(m.Content, "come") || messageContains(m.Content, "coming") || messageContains(m.Content, "cum") {
		if shouldSendMessage() {
			s.ChannelMessageSend(m.ChannelID, "no coming")
		}
	}
}

func Cow(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldIgnoreMessage(s, m) {
		return
	}

	if messageContains(m.Content, "cow") {
		s.ChannelMessageSend(m.ChannelID, "MOOOOO! "+"<:rage:1302882593339084851>")
	}
}

func Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldIgnoreMessage(s, m) {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

func shouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func shouldSendMessage() bool {
	return r.Float64() < chanceToSend
}

func messageContains(str string, substr string) bool {
	toLower := strings.ToLower(str)
	return strings.Contains(toLower, substr)
}
