package advisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/agristream/agristream/internal/models"
)

const anthropicURL = "https://api.anthropic.com/v1/messages"

// Advisor calls the Claude API to generate plain-English farm advice.
type Advisor struct {
	apiKey string
	model  string
	client *http.Client
}

// New constructs an Advisor. If apiKey is empty, Enabled() returns false.
func New(apiKey, model string) *Advisor {
	return &Advisor{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Enabled reports whether the advisor is configured.
func (a *Advisor) Enabled() bool { return a.apiKey != "" }

// anthropicRequest mirrors the Anthropic Messages API request body.
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Advise generates agronomist advice given the current farm state.
func (a *Advisor) Advise(
	ctx context.Context,
	fields []models.Field,
	summaries map[string][]models.SensorReading,
	alerts []models.Alert,
) (string, error) {
	if !a.Enabled() {
		return "", fmt.Errorf("advisor not configured: set ANTHROPIC_API_KEY")
	}

	system := `You are an experienced agronomist assistant for a farm in Kavango East, Namibia.
You receive real-time IoT sensor data from four fields and respond with concise, practical advice.
Focus on immediate actions the farm operator should take. Be specific and brief — 3-5 bullet points maximum.
Do not repeat sensor values back; interpret them and recommend what to do.`

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Current date/time: %s\n\n", time.Now().Format("2006-01-02 15:04 UTC")))
	sb.WriteString(fmt.Sprintf("Active alerts: %d\n\n", len(alerts)))

	for _, f := range fields {
		sb.WriteString(fmt.Sprintf("**%s** (%s, %.1f ha)\n", f.Name, f.CropType, f.Hectares))
		readings := summaries[f.ID]
		if len(readings) == 0 {
			sb.WriteString("  No recent readings\n")
		} else {
			for _, r := range readings {
				sb.WriteString(fmt.Sprintf("  %s: %.2f %s\n", r.Metric, r.Value, r.Unit))
			}
		}
		sb.WriteString("\n")
	}

	if len(alerts) > 0 {
		sb.WriteString("Active alerts:\n")
		for _, al := range alerts {
			sb.WriteString(fmt.Sprintf("  [%s] %s — %s\n", al.Severity, al.Metric, al.Message))
		}
	}

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 512,
		System:    system,
		Messages:  []message{{Role: "user", Content: sb.String()}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, string(body))
	}

	var ar anthropicResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(ar.Content) == 0 || ar.Content[0].Text == "" {
		return "", fmt.Errorf("empty response from anthropic")
	}
	return ar.Content[0].Text, nil
}
