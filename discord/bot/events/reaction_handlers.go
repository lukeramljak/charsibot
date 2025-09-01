package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func SinglePollReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
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
