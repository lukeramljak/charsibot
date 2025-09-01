package events

import "github.com/bwmarrin/discordgo"

func GuildMemberRemove(s *discordgo.Session, gm *discordgo.GuildMemberRemove) {
	s.ChannelMessageSend("1018070065423335437", gm.User.Username+" has left the server. <:periodt:1302882591552307240>")
}
