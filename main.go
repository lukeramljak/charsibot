package main

import (
	events "charsibot/bot"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID = flag.String("guild", "", "Guild ID. If not passed, bot will register commands globally")
	Token   = flag.String("token", "", "Bot access token")
)

var s *discordgo.Session

func init() {
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + *Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	for _, handler := range events.MessageHandlers {
		dg.AddHandler(handler)
	}

	dg.AddHandler(events.GuildMemberRemove)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	defer dg.Close()

	err = dg.UpdateListeningStatus("Big Chungus")
	if err != nil {
		fmt.Println("Error setting listening status: ", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down bot...")
}
