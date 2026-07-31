package server

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/catalog"
	"github.com/lukeramljak/charsibot/db"
	"github.com/lukeramljak/charsibot/stats"
)

func TestAdminRoutesExposeOpenAPI(t *testing.T) {
	mux := http.NewServeMux()
	srv := NewServer(ServerConfig{}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	config := huma.DefaultConfig("Charsibot local admin API", "1.0.0")
	config.DocsPath = "/api/admin/docs"
	config.OpenAPIPath = "/api/admin/openapi"
	api := humago.New(mux, config)
	srv.registerOverlayEvents(api)
	srv.registerAdminRoutes(api)

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/admin/openapi.json", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("OpenAPI status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "blindbox_redemption") {
		t.Fatal("OpenAPI schema does not document the blindbox redemption event")
	}
}

func TestOverlayEventsSendsInitialHeartbeat(t *testing.T) {
	srv := NewServer(ServerConfig{}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	mux := http.NewServeMux()
	srv.NewAPI(mux)
	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(httpServer.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("Content-Type = %q", response.Header.Get("Content-Type"))
	}
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != ": ping\n" {
		t.Fatalf("first SSE line = %q, want initial heartbeat", line)
	}
}

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

func TestAdminResetStatsRestoresDefaults(t *testing.T) {
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
	if err := statsService.ModifyStatValue(t.Context(), "viewer-1", appCatalog.Stats[0].Name, 20); err != nil {
		t.Fatal(err)
	}

	var chatMessage string
	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	srv.SetAdminChatMessage(func(message string) { chatMessage = message })
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/stats/reset", strings.NewReader(`{"displayInChat":true}`))
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	response := httptest.NewRecorder()

	srv.handleAdminResetStats(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	userStats, err := statsService.GetUserStats(t.Context(), "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	for i, stat := range userStats {
		if stat.Value != appCatalog.Stats[i].DefaultValue {
			t.Errorf("%s = %d, want %d", stat.Name, stat.Value, appCatalog.Stats[i].DefaultValue)
		}
	}
	if want := stats.FormatStats("viewer", userStats); chatMessage != want {
		t.Errorf("chat message = %q, want %q", chatMessage, want)
	}
}

func TestAdminExplodeReducesPenisAndDisplaysStats(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	appCatalog, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	for i := range appCatalog.Stats {
		if appCatalog.Stats[i].Name == "penis" {
			appCatalog.Stats[i].DefaultValue = 7
		}
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
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/stats/explode", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	response := httptest.NewRecorder()

	srv.handleAdminExplode(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	userStats, err := statsService.GetUserStats(t.Context(), "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := statValue(userStats, "penis"); got != -1000 {
		t.Errorf("penis stat = %d, want -1000", got)
	}
	if want := stats.FormatStats("viewer", userStats); chatMessage != want {
		t.Errorf("chat message = %q, want %q", chatMessage, want)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/stats/explode/undo", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	response = httptest.NewRecorder()
	srv.handleAdminUndoExplode(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("undo status = %d, body = %s", response.Code, response.Body.String())
	}
	userStats, err = statsService.GetUserStats(t.Context(), "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := statValue(userStats, "penis"); got != 7 {
		t.Errorf("penis stat after undo = %d, want 7", got)
	}
}

func statValue(userStats []stats.UserStat, name string) int64 {
	for _, stat := range userStats {
		if stat.Name == name {
			return stat.Value
		}
	}
	return 0
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

func TestAdminGrantPlushieTriggersRedemptionEvent(t *testing.T) {
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

	series, plushie := appCatalog.Series[0], appCatalog.Series[0].Plushies[0]
	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	request := httptest.NewRequest(http.MethodPut, "/api/admin/users/viewer-1/collections/"+series.Series+"/"+plushie.Key, strings.NewReader(`{"triggerOverlay":true}`))
	request.RemoteAddr = "127.0.0.1:12345"
	request.SetPathValue("userID", "viewer-1")
	request.SetPathValue("series", series.Series)
	request.SetPathValue("key", plushie.Key)
	response := httptest.NewRecorder()
	events := make(chan OverlayEvent, 1)
	srv.clients[events] = struct{}{}

	srv.handleAdminGrantPlushie(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	select {
	case event := <-events:
		data, ok := event.Data.(blindbox.BlindBoxRedemptionData)
		if !ok || event.Type != EventTypeBlindBoxRedemption || data.Plushie.Key != plushie.Key {
			t.Errorf("event = %#v, want redemption for %q", event, plushie.Key)
		}
	default:
		t.Error("expected redemption event")
	}
}

func TestAdminRemovePlushieDoesNotRequireBody(t *testing.T) {
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
	series, plushie := appCatalog.Series[0], appCatalog.Series[0].Plushies[0]
	if _, err := statsService.GetOrCreateStats(t.Context(), "viewer-1", "viewer"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := blindboxService.AddPlushieToCollection(t.Context(), "viewer-1", "viewer", series.Series, plushie.Key); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(ServerConfig{
		StatsService:    statsService,
		BlindBoxService: blindboxService,
		Series:          appCatalog.Series,
	}, slog.New(slog.NewTextHandler(testWriter{t}, nil)))
	mux := http.NewServeMux()
	srv.NewAPI(mux)
	events := make(chan OverlayEvent, 1)
	srv.clients[events] = struct{}{}
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/viewer-1/collections/"+series.Series+"/display", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("display status = %d, body = %s", response.Code, response.Body.String())
	}
	select {
	case event := <-events:
		if event.Type != EventTypeCollectionDisplay {
			t.Errorf("event type = %q, want %q", event.Type, EventTypeCollectionDisplay)
		}
	default:
		t.Error("expected collection display event")
	}
	request = httptest.NewRequest(http.MethodDelete, "/api/admin/users/viewer-1/collections/"+series.Series+"/"+plushie.Key, nil)
	request.RemoteAddr = "127.0.0.1:12345"
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(strings.TrimSpace(string(p)))
	return len(p), nil
}
