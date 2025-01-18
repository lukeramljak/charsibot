package main

import (
	"charsibot/bot"
	"log"
	"os"
	"os/signal"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	token := os.Getenv("TOKEN")

	bot, err := bot.NewBot(appID, guildID, token)
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
