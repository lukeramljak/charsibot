package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var messageHandlers = []func(*discordgo.Session, *discordgo.MessageCreate){
	handleButt,
	handleCome,
	handleCow,
	handleDog,
	handleEgg,
	handleNewMember,
	handlePing,
}

func handleButt(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "but") {
		if chanceSucceeded(0.2) {
			s.ChannelMessageSend(m.ChannelID, "butt")
		}
	}
}

func handleCome(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	comeWords := []string{"come", "coming", "cum", "came"}
	for _, word := range comeWords {
		if messageContains(m.Content, word) {
			if chanceSucceeded(0.2) {
				s.ChannelMessageSend(m.ChannelID, "no coming")
				break
			}
		}
	}
}

func handleCow(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "cow") {
		s.ChannelMessageSend(m.ChannelID, "MOOOOO! "+"<:rage:1302882593339084851>")
	}
}

func handleDog(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "dog") {
		if chanceSucceeded(0.2) {
			s.ChannelMessageSend(m.ChannelID, "what the dog doin'?")
		}
	}
}

func handleEgg(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if messageContains(m.Content, "egg") {
		s.ChannelMessageSend(m.ChannelID, "egg")
	}
}

func handleNewMember(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Type == discordgo.MessageTypeGuildMemberJoin {
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:catJAM:1111234741639848026")
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:hooray:1057490323561001042")
		s.MessageReactionAdd(m.ChannelID, m.ID, "a:pedro:1057490323561001042")
	}
}

func handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isBotMessage(s, m) {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

func handleGuildMemberRemove(s *discordgo.Session, gm *discordgo.GuildMemberRemove) {
	s.ChannelMessageSend("1018070065423335437", gm.User.Username+" has left the server. <:periodt:1302882591552307240>")
}

func handleSinglePollReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		return
	}

	if len(msg.Embeds) == 0 || s.State.User == nil || msg.Author.ID != s.State.User.ID {
		return
	}

	for _, reaction := range msg.Reactions {
		if reaction.Emoji.Name == r.Emoji.Name {
			continue
		}

		err := s.MessageReactionRemove(r.ChannelID, r.MessageID, reaction.Emoji.Name, r.UserID)
		if err != nil {
			log.Printf("Error removing reaction: %v", err)
		}
	}
}

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
