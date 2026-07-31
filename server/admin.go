package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/stats"
)

const explodedPenisValue int64 = -1000

type AdminStat struct {
	Name      string `json:"name" doc:"Stable stat identifier"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
	Value     int64  `json:"value"`
}

type AdminCollection struct {
	Config    blindbox.SeriesConfig `json:"config"`
	Collected []string              `json:"collected" nullable:"false"`
}

type AdminUserResponse struct {
	User        stats.User        `json:"user"`
	Stats       []AdminStat       `json:"stats" nullable:"false"`
	Collections []AdminCollection `json:"collections" nullable:"false"`
}

type adminUsersResponse struct {
	Users []stats.User `json:"users" nullable:"false"`
}
type adminUserOutput struct{ Body AdminUserResponse }
type adminUsersOutput struct{ Body adminUsersResponse }
type adminUserInput struct {
	UserID string `path:"userID" doc:"Twitch user ID"`
}
type adminStatInput struct {
	UserID   string `path:"userID"`
	StatName string `path:"statName"`
	Body     struct {
		Mode  string `json:"mode" enum:"set,adjust"`
		Value int64  `json:"value"`
	}
}
type adminChatInput struct {
	UserID string `path:"userID"`
	Body   struct {
		DisplayInChat bool `json:"displayInChat"`
	}
}
type adminPlushieInput struct {
	UserID string `path:"userID"`
	Series string `path:"series"`
	Key    string `path:"key"`
	Body   struct {
		TriggerOverlay bool `json:"triggerOverlay"`
	}
}
type adminPlushiePathInput struct {
	UserID string `path:"userID"`
	Series string `path:"series"`
	Key    string `path:"key"`
}
type adminRandomPlushieInput struct {
	UserID string `path:"userID"`
	Series string `path:"series"`
	Body   struct {
		TriggerOverlay bool `json:"triggerOverlay"`
	}
}
type adminCollectionInput struct {
	UserID string `path:"userID"`
	Series string `path:"series"`
}

func (s *Server) registerAdminRoutes(api huma.API) {
	admin := huma.NewGroup(api, "/api/admin")
	admin.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		if err := s.requireLocalAdmin(ctx.RemoteAddr()); err != nil {
			status := http.StatusInternalServerError
			if statusErr, ok := err.(huma.StatusError); ok {
				status = statusErr.GetStatus()
			}
			_ = huma.WriteErr(api, ctx, status, "", err)
			return
		}
		next(ctx)
	})

	huma.Register(admin, huma.Operation{OperationID: "list-admin-users", Method: http.MethodGet, Path: "/users", Tags: []string{"Admin"}}, s.listAdminUsers)
	huma.Register(admin, huma.Operation{OperationID: "get-admin-user", Method: http.MethodGet, Path: "/users/{userID}", Tags: []string{"Admin"}}, s.getAdminUser)
	huma.Register(admin, huma.Operation{OperationID: "delete-admin-user", Method: http.MethodDelete, Path: "/users/{userID}", Tags: []string{"Admin"}}, s.deleteAdminUser)
	huma.Register(admin, huma.Operation{OperationID: "update-admin-stat", Method: http.MethodPatch, Path: "/users/{userID}/stats/{statName}", Tags: []string{"Admin"}}, s.updateAdminStat)
	huma.Register(admin, huma.Operation{OperationID: "display-admin-stats", Method: http.MethodPost, Path: "/users/{userID}/stats/display", Tags: []string{"Admin"}}, s.displayAdminStats)
	huma.Register(admin, huma.Operation{OperationID: "grant-admin-random-stat", Method: http.MethodPost, Path: "/users/{userID}/stats/random", Tags: []string{"Admin"}}, s.grantAdminRandomStat)
	huma.Register(admin, huma.Operation{OperationID: "reset-admin-stats", Method: http.MethodPost, Path: "/users/{userID}/stats/reset", Tags: []string{"Admin"}}, s.resetAdminStats)
	huma.Register(admin, huma.Operation{OperationID: "explode-admin-user", Method: http.MethodPost, Path: "/users/{userID}/stats/explode", Tags: []string{"Admin"}}, s.explodeAdminUser)
	huma.Register(admin, huma.Operation{OperationID: "undo-admin-explode", Method: http.MethodPost, Path: "/users/{userID}/stats/explode/undo", Tags: []string{"Admin"}}, s.undoAdminExplode)
	huma.Register(admin, huma.Operation{OperationID: "grant-admin-random-plushie", Method: http.MethodPost, Path: "/users/{userID}/collections/{series}/random", Tags: []string{"Admin"}}, s.grantAdminRandomPlushie)
	huma.Register(admin, huma.Operation{OperationID: "grant-admin-plushie", Method: http.MethodPut, Path: "/users/{userID}/collections/{series}/{key}", Tags: []string{"Admin"}}, s.grantAdminPlushie)
	huma.Register(admin, huma.Operation{OperationID: "remove-admin-plushie", Method: http.MethodDelete, Path: "/users/{userID}/collections/{series}/{key}", Tags: []string{"Admin"}}, s.removeAdminPlushie)
	huma.Register(admin, huma.Operation{OperationID: "display-admin-collection", Method: http.MethodPost, Path: "/users/{userID}/collections/{series}/display", Tags: []string{"Admin"}}, s.displayAdminCollection)
	huma.Register(admin, huma.Operation{OperationID: "reset-admin-collection", Method: http.MethodDelete, Path: "/users/{userID}/collections/{series}", Tags: []string{"Admin"}}, s.resetAdminCollection)
}

func (s *Server) listAdminUsers(ctx context.Context, _ *struct{}) (*adminUsersOutput, error) {
	users, err := s.stats.ListUsers(ctx)
	if err != nil {
		return nil, s.adminError("list users", err)
	}
	return &adminUsersOutput{Body: adminUsersResponse{Users: users}}, nil
}

func (s *Server) getAdminUser(ctx context.Context, input *adminUserInput) (*adminUserOutput, error) {
	return s.adminOutput(ctx, input.UserID)
}

func (s *Server) deleteAdminUser(ctx context.Context, input *adminUserInput) (*struct{}, error) {
	if _, err := s.adminUser(ctx, input.UserID); err != nil {
		return nil, err
	}
	if err := s.stats.DeleteUser(ctx, input.UserID); err != nil {
		return nil, s.adminError("delete user", err)
	}
	return nil, nil
}

func (s *Server) updateAdminStat(ctx context.Context, input *adminStatInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !s.hasStat(input.StatName) {
		return nil, huma.Error400BadRequest("unknown stat")
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	if input.Body.Mode == "set" {
		err = s.stats.SetStatValue(ctx, user.ID, input.StatName, input.Body.Value)
	} else {
		err = s.stats.ModifyStatValue(ctx, user.ID, input.StatName, input.Body.Value)
	}
	if err != nil {
		return nil, s.adminError("update stat", err)
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) displayAdminStats(ctx context.Context, input *adminUserInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureChat(true); err != nil {
		return nil, err
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	if err := s.displayStats(ctx, user, true); err != nil {
		return nil, err
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) grantAdminRandomStat(ctx context.Context, input *adminChatInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureChat(input.Body.DisplayInChat); err != nil {
		return nil, err
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	definition, err := s.stats.GetRandomStatDefinition(ctx)
	if err != nil {
		return nil, s.adminError("choose random stat", err)
	}
	if err := s.stats.ModifyStatValue(ctx, user.ID, definition.Name, 1); err != nil {
		return nil, s.adminError("grant random stat", err)
	}
	if err := s.displayStats(ctx, user, input.Body.DisplayInChat); err != nil {
		return nil, err
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) resetAdminStats(ctx context.Context, input *adminChatInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureChat(input.Body.DisplayInChat); err != nil {
		return nil, err
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	if err := s.stats.ResetStats(ctx, user.ID); err != nil {
		return nil, s.adminError("reset stats", err)
	}
	if err := s.displayStats(ctx, user, input.Body.DisplayInChat); err != nil {
		return nil, err
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) explodeAdminUser(ctx context.Context, input *adminUserInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !s.hasStat("penis") {
		return nil, huma.Error503ServiceUnavailable("penis stat is unavailable")
	}
	if err := s.ensureChat(true); err != nil {
		return nil, err
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	if err := s.stats.SetStatValue(ctx, user.ID, "penis", explodedPenisValue); err != nil {
		return nil, s.adminError("explode stat", err)
	}
	if err := s.displayStats(ctx, user, true); err != nil {
		return nil, err
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) undoAdminExplode(ctx context.Context, input *adminUserInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	penis, found := s.statDefinition("penis")
	if !found {
		return nil, huma.Error503ServiceUnavailable("penis stat is unavailable")
	}
	if err := s.ensureChat(true); err != nil {
		return nil, err
	}
	if _, err := s.stats.GetOrCreateStats(ctx, user.ID, user.Username); err != nil {
		return nil, s.adminError("initialize stats", err)
	}
	if err := s.stats.SetStatValue(ctx, user.ID, "penis", penis.DefaultValue); err != nil {
		return nil, s.adminError("undo explode stat", err)
	}
	if err := s.displayStats(ctx, user, true); err != nil {
		return nil, err
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) grantAdminRandomPlushie(ctx context.Context, input *adminRandomPlushieInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	for _, cfg := range s.series {
		if cfg.Series == input.Series {
			plushie, err := blindbox.PickPlushie(cfg.Plushies)
			if err != nil {
				return nil, s.adminError("choose random plushie", err)
			}
			result, err := s.blindbox.Redeem(ctx, user.ID, user.Username, input.Series, plushie.Key)
			if err != nil {
				return nil, s.adminError("grant random plushie", err)
			}
			if input.Body.TriggerOverlay {
				s.Broadcast(OverlayEvent{Type: EventTypeBlindBoxRedemption, Data: blindbox.BlindBoxRedemptionData{Username: result.Username, Plushie: plushie, IsNew: result.IsNew, Collection: result.Collection, Config: cfg}})
			}
			return s.adminOutput(ctx, user.ID)
		}
	}
	return nil, huma.Error400BadRequest("unknown series")
}

func (s *Server) grantAdminPlushie(ctx context.Context, input *adminPlushieInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	cfg, plushie, found := s.plushieInSeries(input.Series, input.Key)
	if !found {
		return nil, huma.Error400BadRequest("unknown series or plushie")
	}
	isNew, collection, err := s.blindbox.AddPlushieToCollection(ctx, user.ID, user.Username, input.Series, input.Key)
	if err != nil {
		return nil, s.adminError("grant plushie", err)
	}
	if input.Body.TriggerOverlay {
		s.Broadcast(OverlayEvent{Type: EventTypeBlindBoxRedemption, Data: blindbox.BlindBoxRedemptionData{Username: user.Username, Plushie: plushie, IsNew: isNew, Collection: collection, Config: cfg}})
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) removeAdminPlushie(ctx context.Context, input *adminPlushiePathInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !s.hasPlushie(input.Series, input.Key) {
		return nil, huma.Error400BadRequest("unknown series or plushie")
	}
	if err := s.blindbox.RemovePlushieFromCollection(ctx, user.ID, input.Series, input.Key); err != nil {
		return nil, s.adminError("remove plushie", err)
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) resetAdminCollection(ctx context.Context, input *adminCollectionInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !s.hasSeries(input.Series) {
		return nil, huma.Error400BadRequest("unknown series")
	}
	if err := s.blindbox.ResetCollection(ctx, user.ID, input.Series); err != nil {
		return nil, s.adminError("reset collection", err)
	}
	return s.adminOutput(ctx, user.ID)
}

func (s *Server) displayAdminCollection(ctx context.Context, input *adminCollectionInput) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	for _, cfg := range s.series {
		if cfg.Series != input.Series {
			continue
		}
		collection, err := s.blindbox.GetCollection(ctx, user.ID, cfg.Series)
		if err != nil {
			return nil, s.adminError("get collection", err)
		}
		s.Broadcast(OverlayEvent{Type: EventTypeCollectionDisplay, Data: blindbox.BlindBoxDisplayData{Username: user.Username, Collection: collection, Config: cfg}})
		return s.adminOutput(ctx, user.ID)
	}
	return nil, huma.Error400BadRequest("unknown series")
}

func (s *Server) adminOutput(ctx context.Context, userID string) (*adminUserOutput, error) {
	user, err := s.adminUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	userStats, err := s.stats.GetUserStats(ctx, user.ID)
	if err != nil {
		return nil, s.adminError("get stats", err)
	}
	values := make(map[string]int64, len(userStats))
	for _, stat := range userStats {
		values[stat.Name] = stat.Value
	}
	definitions := s.stats.Definitions()
	statValues := make([]AdminStat, len(definitions))
	for i, definition := range definitions {
		value, ok := values[definition.Name]
		if !ok {
			value = definition.DefaultValue
		}
		statValues[i] = AdminStat{Name: definition.Name, ShortName: definition.ShortName, LongName: definition.LongName, Value: value}
	}
	collections := make([]AdminCollection, 0, len(s.series))
	for _, cfg := range s.series {
		collected, err := s.blindbox.GetCollection(ctx, user.ID, cfg.Series)
		if err != nil {
			return nil, s.adminError("get collection", err)
		}
		collections = append(collections, AdminCollection{Config: cfg, Collected: collected})
	}
	return &adminUserOutput{Body: AdminUserResponse{User: user, Stats: statValues, Collections: collections}}, nil
}

func (s *Server) adminUser(ctx context.Context, userID string) (stats.User, error) {
	if strings.TrimSpace(userID) == "" {
		return stats.User{}, huma.Error400BadRequest("user ID is required")
	}
	user, err := s.stats.GetUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return stats.User{}, huma.Error404NotFound("user not found")
	}
	if err != nil {
		return stats.User{}, s.adminError("get user", err)
	}
	return user, nil
}

func (s *Server) ensureChat(display bool) error {
	if display && !s.hasAdminChatMessage() {
		return huma.Error503ServiceUnavailable("chat is unavailable")
	}
	return nil
}
func (s *Server) displayStats(ctx context.Context, user stats.User, display bool) error {
	if !display {
		return nil
	}
	values, err := s.stats.GetUserStats(ctx, user.ID)
	if err != nil {
		return s.adminError("get updated stats", err)
	}
	s.sendAdminChatMessage(stats.FormatStats(user.Username, values))
	return nil
}
func (s *Server) requireLocalAdmin(remoteAddr string) error {
	if s.stats == nil || s.blindbox == nil {
		return huma.Error503ServiceUnavailable("admin is not configured")
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		return huma.Error403Forbidden("admin is available only on localhost")
	}
	return nil
}
func (s *Server) hasAdminChatMessage() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.adminChatMessage != nil
}
func (s *Server) hasStat(name string) bool { _, found := s.statDefinition(name); return found }
func (s *Server) statDefinition(name string) (stats.Definition, bool) {
	for _, stat := range s.stats.Definitions() {
		if stat.Name == name {
			return stat, true
		}
	}
	return stats.Definition{}, false
}
func (s *Server) hasSeries(series string) bool {
	for _, cfg := range s.series {
		if cfg.Series == series {
			return true
		}
	}
	return false
}
func (s *Server) hasPlushie(series, key string) bool {
	_, _, found := s.plushieInSeries(series, key)
	return found
}
func (s *Server) plushieInSeries(series, key string) (blindbox.SeriesConfig, blindbox.Plushie, bool) {
	for _, cfg := range s.series {
		if cfg.Series == series {
			for _, plushie := range cfg.Plushies {
				if plushie.Key == key {
					return cfg, plushie, true
				}
			}
		}
	}
	return blindbox.SeriesConfig{}, blindbox.Plushie{}, false
}
func (s *Server) adminError(action string, err error) error {
	s.logger.Error("admin request failed", slog.String("action", action), slog.Any("err", err))
	return huma.Error500InternalServerError("admin request failed")
}

// The following helpers preserve direct handler coverage while the public API is
// registered through Huma. They are deliberately not registered as routes.
type adminUserResponse = AdminUserResponse

func (s *Server) writeLegacyAdmin(w http.ResponseWriter, r *http.Request, response *adminUserOutput, err error) {
	if err != nil {
		status := http.StatusInternalServerError
		if statusErr, ok := err.(huma.StatusError); ok {
			status = statusErr.GetStatus()
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response.Body)
}

func (s *Server) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	if err := s.requireLocalAdmin(r.RemoteAddr); err != nil {
		s.writeLegacyAdmin(w, r, nil, err)
		return
	}
	out, err := s.getAdminUser(r.Context(), &adminUserInput{UserID: r.PathValue("userID")})
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminRandomStat(w http.ResponseWriter, r *http.Request) {
	input := &adminChatInput{UserID: r.PathValue("userID")}
	_ = json.NewDecoder(r.Body).Decode(&input.Body)
	out, err := s.grantAdminRandomStat(r.Context(), input)
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminResetStats(w http.ResponseWriter, r *http.Request) {
	input := &adminChatInput{UserID: r.PathValue("userID")}
	_ = json.NewDecoder(r.Body).Decode(&input.Body)
	out, err := s.resetAdminStats(r.Context(), input)
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminExplode(w http.ResponseWriter, r *http.Request) {
	out, err := s.explodeAdminUser(r.Context(), &adminUserInput{UserID: r.PathValue("userID")})
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminUndoExplode(w http.ResponseWriter, r *http.Request) {
	out, err := s.undoAdminExplode(r.Context(), &adminUserInput{UserID: r.PathValue("userID")})
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminRandomPlushie(w http.ResponseWriter, r *http.Request) {
	input := &adminRandomPlushieInput{UserID: r.PathValue("userID"), Series: r.PathValue("series")}
	_ = json.NewDecoder(r.Body).Decode(&input.Body)
	out, err := s.grantAdminRandomPlushie(r.Context(), input)
	s.writeLegacyAdmin(w, r, out, err)
}
func (s *Server) handleAdminGrantPlushie(w http.ResponseWriter, r *http.Request) {
	input := &adminPlushieInput{UserID: r.PathValue("userID"), Series: r.PathValue("series"), Key: r.PathValue("key")}
	_ = json.NewDecoder(r.Body).Decode(&input.Body)
	out, err := s.grantAdminPlushie(r.Context(), input)
	s.writeLegacyAdmin(w, r, out, err)
}
