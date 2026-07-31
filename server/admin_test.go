package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/catalog"
	"github.com/lukeramljak/charsibot/db"
	"github.com/lukeramljak/charsibot/stats"
)

func TestAdminUserIncludesDefaultStatsForCollectionOnlyUser(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	appCatalog, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	statsService, err := stats.NewService(queries, appCatalog.Stats)
	if err != nil {
		t.Fatal(err)
	}
	blindboxService, err := blindbox.NewService(queries, appCatalog.Series)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := blindboxService.AddPlushieToCollection(t.Context(), "viewer-1", "viewer", "coobubu", "cutey"); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	request := httptest.NewRequest(http.MethodGet, "/api/admin/users/viewer-1", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	response := httptest.NewRecorder()

	srv.handleAdminUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body adminUserResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Stats) != len(appCatalog.Stats) {
		t.Fatalf("stats = %d, want %d", len(body.Stats), len(appCatalog.Stats))
	}
	if body.Stats[0].Value != appCatalog.Stats[0].DefaultValue {
		t.Fatalf("default stat = %d, want %d", body.Stats[0].Value, appCatalog.Stats[0].DefaultValue)
	}
}

func TestAdminRandomStatIncrementsOneStat(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	appCatalog, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	statsService, err := stats.NewService(queries, appCatalog.Stats)
	if err != nil {
		t.Fatal(err)
	}
	blindboxService, err := blindbox.NewService(queries, appCatalog.Series)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := statsService.GetOrCreateStats(t.Context(), "viewer-1", "viewer"); err != nil {
		t.Fatal(err)
	}

	var chatMessage string
	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	srv.SetAdminChatMessage(func(message string) { chatMessage = message })
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/stats/random", strings.NewReader(`{"displayInChat":true}`))
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	response := httptest.NewRecorder()

	srv.handleAdminRandomStat(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body adminUserResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, stat := range body.Stats {
		total += stat.Value
	}
	var defaults int64
	for _, definition := range appCatalog.Stats {
		defaults += definition.DefaultValue
	}
	if total != defaults+1 {
		t.Errorf("total stat value = %d, want %d", total, defaults+1)
	}
	userStats, err := statsService.GetUserStats(t.Context(), "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	if want := stats.FormatStats("viewer", userStats); chatMessage != want {
		t.Errorf("chat message = %q, want %q", chatMessage, want)
	}
}

func TestAdminRandomPlushieGrantsFromSeries(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	appCatalog, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	statsService, err := stats.NewService(queries, appCatalog.Stats)
	if err != nil {
		t.Fatal(err)
	}
	blindboxService, err := blindbox.NewService(queries, appCatalog.Series)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := statsService.GetOrCreateStats(t.Context(), "viewer-1", "viewer"); err != nil {
		t.Fatal(err)
	}

	series := appCatalog.Series[0]
	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/collections/"+series.Series+"/random", strings.NewReader(`{"triggerOverlay":true}`))
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	request.SetPathValue("series", series.Series)
	response := httptest.NewRecorder()
	events := make(chan OverlayEvent, 1)
	srv.clients[events] = struct{}{}

	srv.handleAdminRandomPlushie(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body adminUserResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Collections) == 0 || len(body.Collections[0].Collected) != 1 {
		t.Fatalf("collected = %#v, want one plushie", body.Collections)
	}
	key := body.Collections[0].Collected[0]
	for _, plushie := range series.Plushies {
		if plushie.Key == key {
			select {
			case event := <-events:
				if event.Type != EventTypeBlindBoxRedemption {
					t.Errorf("event type = %q, want %q", event.Type, EventTypeBlindBoxRedemption)
				}
				return
			default:
				t.Error("expected redemption event")
				return
			}
		}
	}
	t.Errorf("granted plushie %q is not in series %q", key, series.Series)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(strings.TrimSpace(string(p)))
	return len(p), nil
}
