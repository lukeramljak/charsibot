package events

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ShouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func ShouldSendMessage() bool {
	return r.Float64() < chanceToSend
}

func MessageContains(str string, substr string) bool {
	toLower := strings.ToLower(str)
	return strings.Contains(toLower, substr)
}
