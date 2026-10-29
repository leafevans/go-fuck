// Package ai implements the optional AI fallback that is used when no rule
// matched a failed command.
//
// The fallback talks to an OpenAI compatible chat completion endpoint. It is
// disabled by default and must be configured explicitly: sending the failed
// command and its output to a third party is a privacy decision the user has to
// make.
package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/types"
)

// APIKeyEnv is the environment variable that stores the API key. The key is
// deliberately not part of the settings file.
const APIKeyEnv = "GOFUCK_AI_API_KEY"

// fallbackAPIKeyEnv is checked when APIKeyEnv is not set.
const fallbackAPIKeyEnv = "OPENAI_API_KEY"

// maxOutputBytes limits how much of the failed command output is sent.
const maxOutputBytes = 4000

// systemPrompt asks the model for a single shell command.
const systemPrompt = "You fix failed shell commands. Reply with exactly one corrected shell command on a single line, without explanations, quotes or Markdown."

// Client asks an OpenAI compatible endpoint for corrections.
type Client struct {
	// Endpoint is the API base URL, for example https://api.openai.com/v1.
	Endpoint string
	// APIKey authenticates the requests.
	APIKey string
	// Model is the chat completion model name.
	Model string
	// HTTP is the client used for requests.
	HTTP *http.Client
}

// FromSettings returns the client when the AI fallback is enabled and
// configured, or nil.
func FromSettings() *Client {
	settings := conf.Current()
	if !settings.AIEnabled || settings.AIEndpoint == "" || settings.AIModel == "" {
		return nil
	}

	apiKey := os.Getenv(APIKeyEnv)
	if apiKey == "" {
		apiKey = os.Getenv(fallbackAPIKeyEnv)
	}
	if apiKey == "" {
		return nil
	}

	timeout := time.Duration(settings.AITimeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		Endpoint: settings.AIEndpoint,
		APIKey:   apiKey,
		Model:    settings.AIModel,
		HTTP:     &http.Client{Timeout: timeout},
	}
}

// Suggest asks the endpoint for a corrected command. An empty string means
// that the model did not return a usable command.
func (c *Client) Suggest(command *types.Command) (string, error) {
	payload, err := json.Marshal(chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt(command)},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("encode AI request: %w", err)
	}

	url := strings.TrimRight(c.Endpoint, "/") + "/chat/completions"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create AI request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := c.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read AI response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("AI endpoint returned %s: %s", response.Status, truncate(string(body), 200))
	}

	var chatResponse chatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", fmt.Errorf("parse AI response: %w", err)
	}
	if len(chatResponse.Choices) == 0 {
		return "", errors.New("AI response contains no choices")
	}
	return cleanSuggestion(chatResponse.Choices[0].Message.Content), nil
}

// chatRequest is the Chat Completions request body.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// chatMessage is one message of a chat completion.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the subset of the Chat Completions response that is used.
type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// userPrompt builds the user message with the failed command and its output.
func userPrompt(command *types.Command) string {
	output := command.Output
	if len(output) > maxOutputBytes {
		output = output[len(output)-maxOutputBytes:]
	}

	var builder strings.Builder
	builder.WriteString("Command:\n")
	builder.WriteString(command.Script)
	builder.WriteString("\n\nOutput:\n")
	builder.WriteString(output)
	builder.WriteString("\n\nReturn only the corrected command.")
	return builder.String()
}

// cleanSuggestion extracts the first usable command line from the reply.
func cleanSuggestion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "$ ")
		if line == "" || strings.HasPrefix(line, "```") {
			continue
		}
		return line
	}
	return ""
}

// truncate shortens value to at most length bytes.
func truncate(value string, length int) string {
	if len(value) <= length {
		return value
	}
	return value[:length] + "..."
}
