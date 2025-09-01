package main

import (
	"charsibot/bot"
	"log"
	"os"
	"os/signal"
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
