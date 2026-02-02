package blindbox

import (
	"reflect"
	"testing"
)

func TestBuildWeightedPlushieList(t *testing.T) {
	plushies := []PlushieData{
		{Key: 1, Weight: 10},
		{Key: 2, Weight: 8},
		{Key: 3, Weight: 9},
		{Key: 4, Weight: 14},
		{Key: 5, Weight: 1},
		{Key: 6, Weight: 7},
		{Key: 7, Weight: 2},
		{Key: 8, Weight: 10},
	}

	weighted := buildWeightedPlushieList(plushies)

	// Expected weighted list based on weights
	expected := []int{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // reward1 (10x)
		2, 2, 2, 2, 2, 2, 2, 2, // reward2 (8x)
		3, 3, 3, 3, 3, 3, 3, 3, 3, // reward3 (9x)
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, // reward4 (14x)
		5,                   // reward5 (1x)
		6, 6, 6, 6, 6, 6, 6, // reward6 (7x)
		7, 7, // reward7 (2x)
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // reward8 (10x)
	}

	if !reflect.DeepEqual(weighted, expected) {
		t.Errorf("weighted list mismatch\ngot:  %v\nwant: %v", weighted, expected)
	}
}

func buildWeightedPlushieList(plushies []PlushieData) []int {
	weighted := []int{}
	for _, p := range plushies {
		for i := 0; i < p.Weight; i++ {
			weighted = append(weighted, p.Key)
		}
	}
	return weighted
}

func TestGetWeightedRandomPlushie(t *testing.T) {
	plushies := []PlushieData{
		{Key: 1, Weight: 1},
		{Key: 2, Weight: 0},
		{Key: 3, Weight: 0},
	}

	// Since only reward1 has weight, it should always be selected
	for i := 0; i < 10; i++ {
		result := GetWeightedRandomPlushie(plushies)
		if result != 1 {
			t.Errorf("expected reward1 (1), got %d", result)
		}
	}
}

func TestGetWeightedRandomPlushieEmptyList(t *testing.T) {
	plushies := []PlushieData{}
	result := GetWeightedRandomPlushie(plushies)

	// Should return default value of 1
	if result != 1 {
		t.Errorf("expected default value 1, got %d", result)
	}
}

func TestGetWeightedRandomPlushieDistribution(t *testing.T) {
	plushies := []PlushieData{
		{Key: 1, Weight: 100},
		{Key: 2, Weight: 1},
	}

	counts := make(map[int]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		result := GetWeightedRandomPlushie(plushies)
		counts[result]++
	}

	// reward1 should appear much more frequently than reward2
	// With 100:1 ratio, we expect roughly 990:10 distribution
	// Allow some variance but reward1 should be at least 80% of results
	if counts[1] < iterations*80/100 {
		t.Errorf("expected reward1 to appear at least 80%% of the time, got %d/%d (%.1f%%)",
			counts[1], iterations, float64(counts[1])/float64(iterations)*100)
	}
}
