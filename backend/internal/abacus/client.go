package abacus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Client struct {
	apiKey  string
	apiURL  string
	httpCli *http.Client
}

func NewClient(apiKey, apiURL string) *Client {
	return &Client{
		apiKey: apiKey,
		apiURL: apiURL,
		httpCli: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type chatRequest struct {
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (c *Client) Chat(prompt string) (string, error) {
	reqBody := chatRequest{
		Messages: []message{
			{Role: "system", Content: "You are an expert NBA analyst and sports betting advisor. Provide analysis in JSON format when requested."},
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Abacus AI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Abacus AI returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		// If we can't parse the structured response, return raw
		return string(respBody), nil
	}

	if chatResp.Error != "" {
		return "", fmt.Errorf("Abacus AI error: %s", chatResp.Error)
	}

	return chatResp.Response, nil
}

func (c *Client) BuildGameAnalysisPrompt(game *models.NBAGame, homeRecent, awayRecent, h2h []models.NBAGame, odds []models.GameOdds) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Analyze the upcoming NBA game: %s (Home) vs %s (Away) on %s.\n\n",
		game.HomeTeam, game.AwayTeam, game.GameDate.Format("2006-01-02")))

	// Home team recent games
	sb.WriteString(fmt.Sprintf("=== %s Recent Games (Last 10) ===\n", game.HomeTeam))
	for _, g := range homeRecent {
		winner := "W"
		if (g.HomeTeam == game.HomeTeam && g.HomeScore != nil && g.AwayScore != nil && *g.HomeScore < *g.AwayScore) ||
			(g.AwayTeam == game.HomeTeam && g.HomeScore != nil && g.AwayScore != nil && *g.AwayScore < *g.HomeScore) {
			winner = "L"
		}
		homeS, awayS := 0, 0
		if g.HomeScore != nil {
			homeS = *g.HomeScore
		}
		if g.AwayScore != nil {
			awayS = *g.AwayScore
		}
		sb.WriteString(fmt.Sprintf("  %s: %s vs %s — %d-%d (%s)\n", g.GameDate.Format("01/02"), g.HomeTeam, g.AwayTeam, homeS, awayS, winner))
	}

	// Away team recent games
	sb.WriteString(fmt.Sprintf("\n=== %s Recent Games (Last 10) ===\n", game.AwayTeam))
	for _, g := range awayRecent {
		winner := "W"
		if (g.HomeTeam == game.AwayTeam && g.HomeScore != nil && g.AwayScore != nil && *g.HomeScore < *g.AwayScore) ||
			(g.AwayTeam == game.AwayTeam && g.HomeScore != nil && g.AwayScore != nil && *g.AwayScore < *g.HomeScore) {
			winner = "L"
		}
		homeS, awayS := 0, 0
		if g.HomeScore != nil {
			homeS = *g.HomeScore
		}
		if g.AwayScore != nil {
			awayS = *g.AwayScore
		}
		sb.WriteString(fmt.Sprintf("  %s: %s vs %s — %d-%d (%s)\n", g.GameDate.Format("01/02"), g.HomeTeam, g.AwayTeam, homeS, awayS, winner))
	}

	// Head to head
	sb.WriteString("\n=== Head-to-Head (Last 6 Meetings) ===\n")
	for _, g := range h2h {
		homeS, awayS := 0, 0
		if g.HomeScore != nil {
			homeS = *g.HomeScore
		}
		if g.AwayScore != nil {
			awayS = *g.AwayScore
		}
		sb.WriteString(fmt.Sprintf("  %s: %s vs %s — %d-%d\n", g.GameDate.Format("01/02"), g.HomeTeam, g.AwayTeam, homeS, awayS))
	}

	// Odds
	if len(odds) > 0 {
		sb.WriteString("\n=== Current Odds ===\n")
		for _, o := range odds {
			if o.HomeOdd != nil && o.AwayOdd != nil {
				sb.WriteString(fmt.Sprintf("  %s (%s): Home %.2f | Away %.2f\n", o.Bookmaker, o.MarketType, *o.HomeOdd, *o.AwayOdd))
			}
			if o.OverOdd != nil && o.UnderOdd != nil && o.OverUnderLine != nil {
				sb.WriteString(fmt.Sprintf("  O/U %.1f: Over %.2f | Under %.2f\n", *o.OverUnderLine, *o.OverOdd, *o.UnderOdd))
			}
		}
	}

	sb.WriteString(`
Please provide your analysis in the following JSON format:
{
  "win_probability": { "home": 0.00, "away": 0.00 },
  "recommended_bets": [
    { "market": "moneyline|spread|over_under", "pick": "description", "confidence": "high|medium|low", "reasoning": "explanation" }
  ],
  "value_bets": [
    { "market": "type", "pick": "description", "expected_value": 0.00, "reasoning": "explanation" }
  ],
  "risk_level": "low|medium|high",
  "key_factors": ["factor1", "factor2"],
  "summary": "Brief game analysis summary"
}
`)
	return sb.String()
}
