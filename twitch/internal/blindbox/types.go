package blindbox

import (
	"math/rand/v2"
)

type PlushieData struct {
	Key    int
	Weight int
}

type BlindBoxConfig struct {
	CollectionType           string
	RewardTitle              string
	ModeratorCommand         string
	CollectionDisplayCommand string
	ResetCommand             string
	Plushies                 []PlushieData
}

// GetWeightedRandomPlushie returns a random plushie key based on weights
func GetWeightedRandomPlushie(plushies []PlushieData) int {

	weighted := []int{}
	for _, p := range plushies {
		for i := 0; i < p.Weight; i++ {
			weighted = append(weighted, p.Key)
		}
	}

	// Return random selection
	if len(weighted) == 0 {
		return 1 // Default to reward1
	}
	return weighted[rand.IntN(len(weighted))]
}

// Default configurations
var BlindBoxConfigs = []BlindBoxConfig{
	{
		CollectionType:           "coobubu",
		ModeratorCommand:         "coobubu-redeem",
		CollectionDisplayCommand: "coobubu",
		ResetCommand:             "coobubu-reset",
		RewardTitle:              "Cooper Series Blind Box",
		Plushies: []PlushieData{
			{Key: 1, Weight: 12},
			{Key: 2, Weight: 12},
			{Key: 3, Weight: 12},
			{Key: 4, Weight: 12},
			{Key: 5, Weight: 12},
			{Key: 6, Weight: 12},
			{Key: 7, Weight: 12},
			{Key: 8, Weight: 1},
		},
	},
	{
		CollectionType:           "olliepop",
		ModeratorCommand:         "olliepop-redeem",
		CollectionDisplayCommand: "olliepop",
		ResetCommand:             "olliepop-reset",
		RewardTitle:              "Ollie Series Blind Box",
		Plushies: []PlushieData{
			{Key: 1, Weight: 12},
			{Key: 2, Weight: 12},
			{Key: 3, Weight: 12},
			{Key: 4, Weight: 12},
			{Key: 5, Weight: 12},
			{Key: 6, Weight: 12},
			{Key: 7, Weight: 12},
			{Key: 8, Weight: 1},
		},
	},
	{
		CollectionType:           "christmas",
		ModeratorCommand:         "xmas-redeem",
		CollectionDisplayCommand: "xmas",
		ResetCommand:             "xmas-reset",
		RewardTitle:              "Christmas Series Blind Box",
		Plushies: []PlushieData{
			{Key: 1, Weight: 3},
			{Key: 2, Weight: 3},
			{Key: 3, Weight: 3},
			{Key: 4, Weight: 3},
			{Key: 5, Weight: 3},
			{Key: 6, Weight: 3},
			{Key: 7, Weight: 3},
			{Key: 8, Weight: 1},
		},
	},
}
