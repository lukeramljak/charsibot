package events

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func isBotMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func chanceSucceeded(prob float64) bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float64() < prob
}

func messageContains(str string, substr string) bool {
	toLower := strings.ToLower(str)
	return strings.Contains(toLower, substr)
}
