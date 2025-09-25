package stats

type Stat struct {
	Display string
	Column  string
}

type Stats struct {
	ID           string
	Username     string
	Strength     int
	Intelligence int
	Charisma     int
	Luck         int
	Dexterity    int
	Penis        int
}

var statList = []Stat{
	{Display: "Strength", Column: "strength"},
	{Display: "Intelligence", Column: "intelligence"},
	{Display: "Charisma", Column: "charisma"},
	{Display: "Luck", Column: "luck"},
	{Display: "Dexterity", Column: "dexterity"},
	{Display: "Penis", Column: "penis"},
}

func List() []Stat {
	return statList
}
