package catalog

import "testing"

func TestLoadCatalog(t *testing.T) {
	catalog, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(catalog.Stats) == 0 {
		t.Fatal("expected stats")
	}
	if len(catalog.Series) == 0 {
		t.Fatal("expected blind-box series")
	}

	var coobubuFound bool
	var olliepopFound bool
	for _, series := range catalog.Series {
		switch series.Series {
		case "coobubu":
			coobubuFound = true
			if series.RevealSound != "/assets/blind-box/coobubu/reveal.mp3" {
				t.Errorf("coobubu reveal sound = %q", series.RevealSound)
			}
		case "olliepop":
			olliepopFound = true
			if series.RevealSound != "/assets/blind-box/olliepops/reveal.mp3" {
				t.Errorf("olliepop reveal sound = %q", series.RevealSound)
			}
		}
	}
	if !coobubuFound {
		t.Error("coobubu series missing")
	}
	if !olliepopFound {
		t.Error("olliepop series missing")
	}
}

func TestSeriesConfigValidation(t *testing.T) {
	tests := []struct {
		name string
		cfg  seriesJSON
	}{
		{
			name: "requires a plushie",
			cfg:  seriesJSON{Series: "test", RedemptionTitle: "Test", Name: "Tests"},
		},
		{
			name: "requires positive plushie weight",
			cfg: seriesJSON{
				Series: "test", RedemptionTitle: "Test", Name: "Tests",
				Plushies: []plushieJSON{{Key: "one", Weight: 0}},
			},
		},
		{
			name: "requires unique plushie sort orders",
			cfg: seriesJSON{
				Series: "test", RedemptionTitle: "Test", Name: "Tests",
				Plushies: []plushieJSON{{Key: "one", Weight: 1}, {Key: "two", Weight: 1}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.cfg.toSeriesConfig(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
