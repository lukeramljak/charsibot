package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/stats"
)

type adminStat struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
	Value     int64  `json:"value"`
}

type adminCollection struct {
	Config    blindbox.SeriesConfig `json:"config"`
	Collected []string              `json:"collected"`
}

type adminUserResponse struct {
	User        stats.User        `json:"user"`
	Stats       []adminStat       `json:"stats"`
	Collections []adminCollection `json:"collections"`
}

type updateStatRequest struct {
	Mode  string `json:"mode"`
	Value int64  `json:"value"`
}

type randomPlushieRequest struct {
	TriggerOverlay bool `json:"triggerOverlay"`
}

type randomStatRequest struct {
	DisplayInChat bool `json:"displayInChat"`
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	users, err := s.stats.ListUsers(r.Context())
	if err != nil {
		s.adminError(w, "list users", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	s.writeAdminUser(w, r, r.PathValue("userID"))
}

func (s *Server) handleAdminStat(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	statName := r.PathValue("statName")
	if !s.hasStat(statName) {
		http.Error(w, "unknown stat", http.StatusBadRequest)
		return
	}
	var input updateStatRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if _, err := s.stats.GetOrCreateStats(r.Context(), user.ID, user.Username); err != nil {
		s.adminError(w, "initialize stats", err)
		return
	}
	var err error
	switch input.Mode {
	case "set":
		err = s.stats.SetStatValue(r.Context(), user.ID, statName, input.Value)
	case "adjust":
		err = s.stats.ModifyStatValue(r.Context(), user.ID, statName, input.Value)
	default:
		http.Error(w, `mode must be "set" or "adjust"`, http.StatusBadRequest)
		return
	}
	if err != nil {
		s.adminError(w, "update stat", err)
		return
	}
	s.writeAdminUser(w, r, user.ID)
}

func (s *Server) handleAdminRandomStat(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	var input randomStatRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.DisplayInChat && !s.hasAdminChatMessage() {
		http.Error(w, "chat is unavailable", http.StatusServiceUnavailable)
		return
	}
	if _, err := s.stats.GetOrCreateStats(r.Context(), user.ID, user.Username); err != nil {
		s.adminError(w, "initialize stats", err)
		return
	}
	definition, err := s.stats.GetRandomStatDefinition(r.Context())
	if err != nil {
		s.adminError(w, "choose random stat", err)
		return
	}
	if err := s.stats.ModifyStatValue(r.Context(), user.ID, definition.Name, 1); err != nil {
		s.adminError(w, "grant random stat", err)
		return
	}
	if input.DisplayInChat {
		userStats, err := s.stats.GetUserStats(r.Context(), user.ID)
		if err != nil {
			s.adminError(w, "get updated stats", err)
			return
		}
		s.sendAdminChatMessage(stats.FormatStats(user.Username, userStats))
	}
	s.writeAdminUser(w, r, user.ID)
}

func (s *Server) handleAdminRandomPlushie(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	var input randomPlushieRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	series := r.PathValue("series")
	for _, cfg := range s.series {
		if cfg.Series != series {
			continue
		}
		plushie, err := blindbox.PickPlushie(cfg.Plushies)
		if err != nil {
			s.adminError(w, "choose random plushie", err)
			return
		}
		result, err := s.blindbox.Redeem(r.Context(), user.ID, user.Username, series, plushie.Key)
		if err != nil {
			s.adminError(w, "grant random plushie", err)
			return
		}
		if input.TriggerOverlay {
			s.Broadcast(OverlayEvent{
				Type: EventTypeBlindBoxRedemption,
				Data: blindbox.BlindBoxRedemptionData{
					Username:   result.Username,
					Plushie:    plushie,
					IsNew:      result.IsNew,
					Collection: result.Collection,
					Config:     cfg,
				},
			})
		}
		s.writeAdminUser(w, r, user.ID)
		return
	}
	http.Error(w, "unknown series", http.StatusBadRequest)
}

func (s *Server) handleAdminGrantPlushie(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	series, key := r.PathValue("series"), r.PathValue("key")
	if !s.hasPlushie(series, key) {
		http.Error(w, "unknown series or plushie", http.StatusBadRequest)
		return
	}
	if _, _, err := s.blindbox.AddPlushieToCollection(r.Context(), user.ID, user.Username, series, key); err != nil {
		s.adminError(w, "grant plushie", err)
		return
	}
	s.writeAdminUser(w, r, user.ID)
}

func (s *Server) handleAdminDeletePlushie(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	series, key := r.PathValue("series"), r.PathValue("key")
	if !s.hasPlushie(series, key) {
		http.Error(w, "unknown series or plushie", http.StatusBadRequest)
		return
	}
	if err := s.blindbox.RemovePlushieFromCollection(r.Context(), user.ID, series, key); err != nil {
		s.adminError(w, "remove plushie", err)
		return
	}
	s.writeAdminUser(w, r, user.ID)
}

func (s *Server) handleAdminResetCollection(w http.ResponseWriter, r *http.Request) {
	if !s.requireLocalAdmin(w, r) {
		return
	}
	user, ok := s.adminUser(w, r)
	if !ok {
		return
	}
	series := r.PathValue("series")
	if !s.hasSeries(series) {
		http.Error(w, "unknown series", http.StatusBadRequest)
		return
	}
	if err := s.blindbox.ResetCollection(r.Context(), user.ID, series); err != nil {
		s.adminError(w, "reset collection", err)
		return
	}
	s.writeAdminUser(w, r, user.ID)
}

func (s *Server) writeAdminUser(w http.ResponseWriter, r *http.Request, userID string) {
	user, ok := s.adminUserByID(w, r, userID)
	if !ok {
		return
	}
	userStats, err := s.stats.GetUserStats(r.Context(), user.ID)
	if err != nil {
		s.adminError(w, "get stats", err)
		return
	}
	valuesByName := make(map[string]int64, len(userStats))
	for _, stat := range userStats {
		valuesByName[stat.Name] = stat.Value
	}
	definitions := s.stats.Definitions()
	statValues := make([]adminStat, len(definitions))
	for i, definition := range definitions {
		value, exists := valuesByName[definition.Name]
		if !exists {
			value = definition.DefaultValue
		}
		statValues[i] = adminStat{
			Name:      definition.Name,
			ShortName: definition.ShortName,
			LongName:  definition.LongName,
			Value:     value,
		}
	}
	collections := make([]adminCollection, 0, len(s.series))
	for _, cfg := range s.series {
		collected, err := s.blindbox.GetCollection(r.Context(), user.ID, cfg.Series)
		if err != nil {
			s.adminError(w, "get collection", err)
			return
		}
		collections = append(collections, adminCollection{Config: cfg, Collected: collected})
	}
	writeJSON(w, http.StatusOK, adminUserResponse{User: user, Stats: statValues, Collections: collections})
}

func (s *Server) adminUser(w http.ResponseWriter, r *http.Request) (stats.User, bool) {
	return s.adminUserByID(w, r, r.PathValue("userID"))
}

func (s *Server) adminUserByID(w http.ResponseWriter, r *http.Request, userID string) (stats.User, bool) {
	if strings.TrimSpace(userID) == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return stats.User{}, false
	}
	user, err := s.stats.GetUser(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "user not found", http.StatusNotFound)
		return stats.User{}, false
	}
	if err != nil {
		s.adminError(w, "get user", err)
		return stats.User{}, false
	}
	return user, true
}

func (s *Server) requireLocalAdmin(w http.ResponseWriter, r *http.Request) bool {
	if s.stats == nil || s.blindbox == nil {
		http.Error(w, "admin is not configured", http.StatusServiceUnavailable)
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		http.Error(w, "admin is available only on localhost", http.StatusForbidden)
		return false
	}
	return true
}

func (s *Server) hasAdminChatMessage() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.adminChatMessage != nil
}

func (s *Server) hasStat(name string) bool {
	for _, stat := range s.stats.Definitions() {
		if stat.Name == name {
			return true
		}
	}
	return false
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
	for _, cfg := range s.series {
		if cfg.Series != series {
			continue
		}
		for _, plushie := range cfg.Plushies {
			if plushie.Key == key {
				return true
			}
		}
	}
	return false
}

func (s *Server) adminError(w http.ResponseWriter, action string, err error) {
	s.logger.Error("admin request failed", "action", action, "err", err)
	http.Error(w, "admin request failed", http.StatusInternalServerError)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid JSON request body", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
