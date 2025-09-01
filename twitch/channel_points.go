package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
)

func (tc *TwitchClient) handleChannelPointsRedemption(event *Event) error {
	if event.Reward == nil {
		slog.Warn("Received channel points redemption with no reward data")
		return nil
	}

	slog.Info("🎯 Channel points redeemed",
		"user", event.UserLogin,
		"reward", event.Reward.Title,
		"cost", event.Reward.Cost,
		"user_input", event.UserInput)

	switch event.Reward.Title {
	case "Drink a Potion":
		return tc.handleIncreaseRandomStat(event.UserLogin)

	case "Tempt the Dice":
		return tc.handleRollDice(event.UserLogin)

	default:
		return nil
	}
}

func (tc *TwitchClient) handleIncreaseRandomStat(username string) error {
	randomStat := tc.getRandomStat()
	randomModifier := tc.randomStatModifier()
	statLower := strings.ToLower(randomStat)

	updateSQL := fmt.Sprintf("UPDATE stats SET %s = %s + ? WHERE username = ?", statLower, statLower)

	result, err := tc.db.Exec(updateSQL, randomModifier, username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err := tc.db.Exec("INSERT INTO stats (username) VALUES (?)", username)
		if err != nil {
			return fmt.Errorf("failed to create stats for user %s: %w", username, err)
		}
		slog.Info("Created new stats for user", "username", username)

		_, err = tc.db.Exec(updateSQL, randomModifier, username)
		if err != nil {
			return err
		}
	}

	outcome := "gained"
	if randomModifier < 0 {
		outcome = "lost"
	}

	message := fmt.Sprintf(
		"A shifty looking merchant hands @%s a glittering potion. "+
			"Without hesitation, they sink the whole drink. "+
			"@%s %s %s",
		username, username, outcome, randomStat)

	return tc.SendChatMessage(message)
}

func (tc *TwitchClient) handleRollDice(username string) error {
	tc.SendChatMessage(fmt.Sprintf("@%s has rolled with initiative.", username))
	return tc.handleStats(username, "")
}

func (tc *TwitchClient) getRandomStat() string {
	stats := []string{"Strength", "Intelligence", "Charisma", "Luck", "Dexterity", "Penis"}
	return stats[rand.Intn(len(stats))]
}

func (tc *TwitchClient) randomStatModifier() int {
	if rand.Intn(20) == 0 {
		return -1
	}
	return 1
}
