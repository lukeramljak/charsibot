package main

import (
	"charsibot/bot"
	"flag"
	"log"
	"os"
	"os/signal"
)

var (
	GuildID = flag.String("guild", "", "Guild ID. If not passed, bot will register commands globally")
	Token   = flag.String("token", "", "Bot access token")
)

func init() {
	flag.Parse()
}

func main() {
	bot, err := bot.NewBot(*Token)
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
