package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/lukeramljak/charsibot/discord/bot"
)

func main() {
	bot, err := bot.NewBot()
	if err != nil {
		log.Fatal(err)
	}

	bot.RegisterHandlers()
	err = bot.RegisterCommands()
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
