package events

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func ShouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func ShouldSendMessage() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float64() < 0.2
}

func MessageContains(str string, substr string) bool {
	toLower := strings.ToLower(str)
	return strings.Contains(toLower, substr)
}
