package main

import (
	"charsibot/bot"
	"flag"
	"log"
	"os"
	"os/signal"
)

var (
	appID   = flag.String("app", "", "Application ID")
	guildID = flag.String("guild", "", "Guild ID. If not passed, bot will register commands globally")
	token   = flag.String("token", "", "Bot access token")
)

func init() {
	flag.Parse()
}

func main() {
	bot, err := bot.NewBot(*appID, *guildID, *token)
	if err != nil {
		log.Fatal(err)
	}

	go bot.Start()
	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down bot...")
}
