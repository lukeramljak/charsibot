package blindbox

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/server"
	"github.com/lukeramljak/charsibot/internal/store"
)

// RedeemCommand allows moderators to manually trigger a blind box redemption
type RedeemCommand struct {
	config BlindBoxConfig
}

func NewRedeemCommand(config BlindBoxConfig) *RedeemCommand {
	return &RedeemCommand{config: config}
}

func (c *RedeemCommand) ModeratorOnly() bool {
	return true
}

func (c *RedeemCommand) ShouldTrigger(command string) bool {
	return command == c.config.ModeratorCommand
}

func (c *RedeemCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	RedeemBlindBox(b, event.ChatterUserId, event.ChatterUserName, c.config)
}

// ShowCollectionCommand displays a user's blind box collection on the overlay
type ShowCollectionCommand struct {
	config BlindBoxConfig
}

func NewShowCollectionCommand(config BlindBoxConfig) *ShowCollectionCommand {
	return &ShowCollectionCommand{config: config}
}

func (c *ShowCollectionCommand) ModeratorOnly() bool {
	return false
}

func (c *ShowCollectionCommand) ShouldTrigger(command string) bool {
	return command == c.config.CollectionDisplayCommand
}

func (c *ShowCollectionCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	userId := event.ChatterUserId
	username := event.ChatterUserName

	uc, err := b.Store().GetUserCollectionRow(b.Context(), store.GetUserCollectionRowParams{
		UserID:         sql.NullString{String: userId, Valid: true},
		CollectionType: sql.NullString{String: c.config.CollectionType, Valid: true},
	})

	var collection []int
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			collection = []int{}
		} else {
			slog.Error("failed to get collection", "err", err, "user", username)
			b.SendMessage(bot.SendMessageParams{
				Message: fmt.Sprintf("Failed to get %s's collection", username),
			})
			return
		}
	} else {
		collection = store.GetUserCollection(uc)
	}

	b.BroadcastOverlayEvent(server.OverlayEvent{
		Type: "collection_display",
		Data: map[string]any{
			"userId":         userId,
			"username":       username,
			"collectionType": c.config.CollectionType,
			"collection":     intSliceToRewardKeys(collection),
			"collectionSize": len(collection),
		},
	})

	slog.Info("displaying collection", "user", username, "collection", c.config.CollectionType, "size", len(collection))
}

// ResetCommand allows moderators to reset their blind box collection
type ResetCommand struct {
	config BlindBoxConfig
}

func NewResetCommand(config BlindBoxConfig) *ResetCommand {
	return &ResetCommand{config: config}
}

func (c *ResetCommand) ModeratorOnly() bool {
	return true
}

func (c *ResetCommand) ShouldTrigger(command string) bool {
	return command == c.config.ResetCommand
}

func (c *ResetCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	userId := event.ChatterUserId

	err := b.Store().ResetUserCollection(b.Context(), store.ResetUserCollectionParams{
		UserID:         sql.NullString{String: userId, Valid: true},
		CollectionType: sql.NullString{String: c.config.CollectionType, Valid: true},
	})

	if err != nil {
		slog.Error("failed to reset collection", "err", err, "user", event.ChatterUserName)
		return
	}

	slog.Info("collection reset", "user", event.ChatterUserName, "collection", c.config.CollectionType)
}

// CompletedCollectionsCommand displays all users who have completed each collection
type CompletedCollectionsCommand struct{}

func NewCompletedCollectionsCommand() *CompletedCollectionsCommand {
	return &CompletedCollectionsCommand{}
}

func (c *CompletedCollectionsCommand) ModeratorOnly() bool {
	return false
}

func (c *CompletedCollectionsCommand) ShouldTrigger(command string) bool {
	return command == "collections"
}

func (c *CompletedCollectionsCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	collections, err := b.Store().GetCompletedCollections(b.Context())
	if err != nil {
		slog.Error("failed to get completed collections", "err", err)
		return
	}

	// Send header
	b.SendMessage(bot.SendMessageParams{
		Message: "The following chatters have completed the below blind box collections:",
	})

	// Send each collection
	for _, collection := range collections {
		collectionType := collection.CollectionType.String
		if collectionType != "" {
			collectionType = strings.ToUpper(collectionType[:1]) + collectionType[1:]
		}

		usernames := strings.Split(collection.UsernamesCsv, ",")
		message := fmt.Sprintf("%s: %s", collectionType, strings.Join(usernames, ", "))

		b.SendMessage(bot.SendMessageParams{
			Message: message,
		})
	}
}
