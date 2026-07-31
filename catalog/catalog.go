package catalog

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/stats"
)

//go:embed config/stats.json config/blind-box/*.json
var files embed.FS

type Catalog struct {
	Stats  []stats.Definition
	Series []blindbox.SeriesConfig
}

type statDefinitionJSON struct {
	Name         string `json:"name"`
	ShortName    string `json:"shortName"`
	LongName     string `json:"longName"`
	DefaultValue int64  `json:"defaultValue"`
	SortOrder    int64  `json:"sortOrder"`
	Emoji        string `json:"emoji"`
}

type seriesJSON struct {
	Series          string        `json:"series"`
	AssetDir        string        `json:"assetDir"`
	RedemptionTitle string        `json:"redemptionTitle"`
	Name            string        `json:"name"`
	RevealSound     string        `json:"revealSound"`
	BoxFrontFace    string        `json:"boxFrontFace"`
	BoxSideFace     string        `json:"boxSideFace"`
	DisplayColor    string        `json:"displayColor"`
	TextColor       string        `json:"textColor"`
	Plushies        []plushieJSON `json:"plushies"`
}

type plushieJSON struct {
	Key        string `json:"key"`
	SortOrder  int64  `json:"sortOrder"`
	Weight     int64  `json:"weight"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	EmptyImage string `json:"emptyImage"`
}

func Load() (Catalog, error) {
	stats, err := loadStats()
	if err != nil {
		return Catalog{}, err
	}
	series, err := loadSeries()
	if err != nil {
		return Catalog{}, err
	}
	return Catalog{Stats: stats, Series: series}, nil
}

func loadStats() ([]stats.Definition, error) {
	var raw []statDefinitionJSON
	if err := decodeJSON("config/stats.json", &raw); err != nil {
		return nil, err
	}

	definitions := make([]stats.Definition, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	seenSortOrders := make(map[int64]struct{}, len(raw))

	for _, stat := range raw {
		if strings.TrimSpace(stat.Name) == "" ||
			strings.TrimSpace(stat.ShortName) == "" ||
			strings.TrimSpace(stat.LongName) == "" {
			return nil, errors.New("stat name, shortName, and longName are required")
		}
		if _, ok := seen[stat.Name]; ok {
			return nil, fmt.Errorf("duplicate stat %q", stat.Name)
		}
		if _, ok := seenSortOrders[stat.SortOrder]; ok {
			return nil, fmt.Errorf("duplicate stat sortOrder %d", stat.SortOrder)
		}

		seen[stat.Name] = struct{}{}
		seenSortOrders[stat.SortOrder] = struct{}{}

		definitions = append(definitions, stats.Definition{
			Name:         stat.Name,
			ShortName:    stat.ShortName,
			LongName:     stat.LongName,
			DefaultValue: stat.DefaultValue,
			SortOrder:    stat.SortOrder,
			Emoji:        stat.Emoji,
		})
	}
	if len(definitions) == 0 {
		return nil, errors.New("at least one stat is required")
	}

	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].SortOrder < definitions[j].SortOrder
	})

	return definitions, nil
}

func loadSeries() ([]blindbox.SeriesConfig, error) {
	entries, err := fs.ReadDir(files, "config/blind-box")
	if err != nil {
		return nil, fmt.Errorf("read blind-box config dir: %w", err)
	}

	series := make([]blindbox.SeriesConfig, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var raw seriesJSON
		if err := decodeJSON(path.Join("config/blind-box", entry.Name()), &raw); err != nil {
			return nil, err
		}
		cfg, err := raw.toSeriesConfig()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if _, ok := seen[cfg.Series]; ok {
			return nil, fmt.Errorf("duplicate blind-box series %q", cfg.Series)
		}
		seen[cfg.Series] = struct{}{}
		series = append(series, cfg)
	}
	if len(series) == 0 {
		return nil, errors.New("at least one blind-box series is required")
	}

	sort.Slice(series, func(i, j int) bool {
		return series[i].Series < series[j].Series
	})

	return series, nil
}

func (s seriesJSON) toSeriesConfig() (blindbox.SeriesConfig, error) {
	if strings.TrimSpace(s.Series) == "" {
		return blindbox.SeriesConfig{}, errors.New("series is required")
	}
	if strings.TrimSpace(s.RedemptionTitle) == "" || strings.TrimSpace(s.Name) == "" {
		return blindbox.SeriesConfig{}, errors.New("redemptionTitle and name are required")
	}
	if len(s.Plushies) == 0 {
		return blindbox.SeriesConfig{}, errors.New("at least one plushie is required")
	}

	assetDir := s.AssetDir
	if assetDir == "" {
		assetDir = s.Series
	}

	cfg := blindbox.SeriesConfig{
		Series:          s.Series,
		RedemptionTitle: s.RedemptionTitle,
		Name:            s.Name,
		RevealSound:     assetURL(assetDir, s.RevealSound),
		BoxFrontFace:    assetURL(assetDir, s.BoxFrontFace),
		BoxSideFace:     assetURL(assetDir, s.BoxSideFace),
		DisplayColor:    s.DisplayColor,
		TextColor:       s.TextColor,
		Plushies:        make([]blindbox.Plushie, 0, len(s.Plushies)),
	}

	seen := make(map[string]struct{}, len(s.Plushies))
	seenSortOrders := make(map[int64]struct{}, len(s.Plushies))

	for _, plushie := range s.Plushies {
		if strings.TrimSpace(plushie.Key) == "" {
			return blindbox.SeriesConfig{}, errors.New("plushie key is required")
		}
		if plushie.Weight <= 0 {
			return blindbox.SeriesConfig{}, fmt.Errorf("plushie %q must have a positive weight", plushie.Key)
		}
		if _, ok := seen[plushie.Key]; ok {
			return blindbox.SeriesConfig{}, fmt.Errorf("duplicate plushie %q", plushie.Key)
		}
		if _, ok := seenSortOrders[plushie.SortOrder]; ok {
			return blindbox.SeriesConfig{}, fmt.Errorf("duplicate plushie sortOrder %d", plushie.SortOrder)
		}

		seen[plushie.Key] = struct{}{}
		seenSortOrders[plushie.SortOrder] = struct{}{}

		cfg.Plushies = append(cfg.Plushies, blindbox.Plushie{
			Series:     s.Series,
			Key:        plushie.Key,
			SortOrder:  plushie.SortOrder,
			Weight:     plushie.Weight,
			Name:       plushie.Name,
			Image:      assetURL(assetDir, plushie.Image),
			EmptyImage: assetURL(assetDir, plushie.EmptyImage),
		})
	}

	sort.Slice(cfg.Plushies, func(i, j int) bool {
		return cfg.Plushies[i].SortOrder < cfg.Plushies[j].SortOrder
	})

	return cfg, nil
}

func assetURL(assetDir, filename string) string {
	if strings.HasPrefix(filename, "/") {
		return filename
	}
	return path.Join("/assets/blind-box", assetDir, filename)
}

func decodeJSON(name string, dest any) error {
	file, err := files.Open(name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode %s: expected a single JSON value", name)
	}

	return nil
}
