package stats_test

import (
	"testing"

	"github.com/lukeramljak/charsibot/twitch/db"
	"github.com/lukeramljak/charsibot/twitch/stats"
)

func TestFormatStats(t *testing.T) {
	userStats := []db.GetUserStatsRow{
		{Name: "strength", ShortName: "STR", LongName: "Strength", Value: 5},
		{Name: "intelligence", ShortName: "INT", LongName: "Intelligence", Value: 5},
		{Name: "charisma", ShortName: "CHA", LongName: "Charisma", Value: 3},
		{Name: "luck", ShortName: "LUCK", LongName: "Luck", Value: 3},
		{Name: "dexterity", ShortName: "DEX", LongName: "Dexterity", Value: 3},
		{Name: "penis", ShortName: "PENIS", LongName: "Penis", Value: 3},
	}

	formatted := stats.FormatStats("testuser", userStats)
	expected := "testuser's stats: STR: 5 | INT: 5 | CHA: 3 | LUCK: 3 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

func TestFormatStatsNegative(t *testing.T) {
	userStats := []db.GetUserStatsRow{
		{Name: "strength", ShortName: "STR", LongName: "Strength", Value: 3},
		{Name: "intelligence", ShortName: "INT", LongName: "Intelligence", Value: 3},
		{Name: "charisma", ShortName: "CHA", LongName: "Charisma", Value: 9},
		{Name: "luck", ShortName: "LUCK", LongName: "Luck", Value: -2},
		{Name: "dexterity", ShortName: "DEX", LongName: "Dexterity", Value: 3},
		{Name: "penis", ShortName: "PENIS", LongName: "Penis", Value: 3},
	}

	formatted := stats.FormatStats("testuser", userStats)
	expected := "testuser's stats: STR: 3 | INT: 3 | CHA: 9 | LUCK: -2 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

func TestFormatStatsEmpty(t *testing.T) {
	formatted := stats.FormatStats("testuser", []db.GetUserStatsRow{})
	expected := "testuser's stats: "

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}
