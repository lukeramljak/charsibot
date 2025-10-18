package main

import (
	"log"
	"os"
	"os/signal"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	bot, err := NewBot()
	if err != nil {
		log.Fatal(err)
	}

	bot.RegisterHandlers()
	if err := bot.RegisterCommands(); err != nil {
		log.Fatal(err)
	}

	go bot.Start()
	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Shutting down bot...")
}
