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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model,omitempty"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature"`
}

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Chat(prompt string) (string, error) {
	reqBody := chatCompletionRequest{
		Messages: []chatMessage{
			{Role: "system", Content: "You are an expert NBA analyst and sports betting advisor. Provide analysis in JSON format when requested."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/chat/completions", bytes.NewReader(body))
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

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return string(respBody), nil
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("Abacus AI error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("Abacus AI returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
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
You MUST provide your analysis in the following JSON format. You MUST include EXACTLY 3 recommended bets — one for each market: moneyline, over_under, and handicap. Do NOT omit any of the 3 markets.
{
  "win_probability": { "home": 0.00, "away": 0.00 },
  "recommended_bets": [
    { "market": "moneyline", "pick": "Team Name", "confidence": "high|medium|low", "reasoning": "detailed explanation of why this team wins" },
    { "market": "over_under", "pick": "Over/Under X.X (assuming line is around this)", "confidence": "high|medium|low", "reasoning": "detailed explanation based on recent scoring trends" },
    { "market": "handicap", "pick": "Team Name -X.X", "confidence": "high|medium|low", "reasoning": "detailed explanation of expected margin" }
  ],
  "risk_level": "low|medium|high",
  "key_factors": ["factor1", "factor2", "factor3"],
  "summary": "Comprehensive game analysis summary with betting outlook"
}
`)
	return sb.String()
}
