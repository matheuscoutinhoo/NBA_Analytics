package unit

import (
	"testing"
	"time"

	"github.com/matheuscoutinhoo/better/internal/abacus"
	"github.com/matheuscoutinhoo/better/internal/models"
)

func TestBuildGameAnalysisPrompt(t *testing.T) {
	client := abacus.NewClient("test-key", "http://localhost")

	homeScore := 110
	awayScore := 105
	game := &models.NBAGame{
		ID:       1,
		HomeTeam: "Los Angeles Lakers",
		AwayTeam: "Boston Celtics",
		GameDate: time.Now().AddDate(0, 0, 1),
	}

	homeRecent := []models.NBAGame{
		{HomeTeam: "Los Angeles Lakers", AwayTeam: "Denver Nuggets", HomeScore: &homeScore, AwayScore: &awayScore, GameDate: time.Now().AddDate(0, 0, -1)},
	}

	awayRecent := []models.NBAGame{
		{HomeTeam: "Boston Celtics", AwayTeam: "Miami Heat", HomeScore: &homeScore, AwayScore: &awayScore, GameDate: time.Now().AddDate(0, 0, -2)},
	}

	h2h := []models.NBAGame{
		{HomeTeam: "Los Angeles Lakers", AwayTeam: "Boston Celtics", HomeScore: &homeScore, AwayScore: &awayScore, GameDate: time.Now().AddDate(0, 0, -30)},
	}

	homeOdd := 1.85
	awayOdd := 2.10
	odds := []models.GameOdds{
		{GameID: 1, Bookmaker: "test", MarketType: "h2h", HomeOdd: &homeOdd, AwayOdd: &awayOdd},
	}

	prompt := client.BuildGameAnalysisPrompt(game, homeRecent, awayRecent, h2h, odds)

	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	// Check key elements are present
	if !contains(prompt, "Los Angeles Lakers") {
		t.Error("prompt should contain home team name")
	}
	if !contains(prompt, "Boston Celtics") {
		t.Error("prompt should contain away team name")
	}
	if !contains(prompt, "Recent Games") {
		t.Error("prompt should contain recent games section")
	}
	if !contains(prompt, "Head-to-Head") {
		t.Error("prompt should contain head-to-head section")
	}
	if !contains(prompt, "Current Odds") {
		t.Error("prompt should contain odds section")
	}
	if !contains(prompt, "win_probability") {
		t.Error("prompt should contain expected JSON format")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		len(s) >= len(substr) &&
		containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
