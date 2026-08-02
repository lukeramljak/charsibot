// Command openapi prints the generated Charsibot OpenAPI document as JSON.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/lukeramljak/charsibot/server"
)

func main() {
	api := server.NewServer(server.ServerConfig{}, slog.New(slog.NewTextHandler(os.Stderr, nil))).
		NewAPI(http.NewServeMux())
	document, err := api.OpenAPI().MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("marshal OpenAPI document: %w", err))
	}
	if _, err := os.Stdout.Write(document); err != nil {
		panic(fmt.Errorf("write OpenAPI document: %w", err))
	}
}
