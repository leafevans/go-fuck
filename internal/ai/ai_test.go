package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/types"
)

// codeFence returns three backticks without using a Go raw string.
func codeFence() string {
	return strings.Repeat(string(rune(96)), 3)
}

// setAISettings enables the fallback with the given endpoint and model.
func setAISettings(t *testing.T, endpoint, model string) {
	t.Helper()
	conf.Set(&conf.Settings{
		AIEnabled:  true,
		AIEndpoint: endpoint,
		AIModel:    model,
		AITimeout:  5,
	})
	t.Cleanup(conf.Reset)
}

// writeChatResponse writes an OpenAI compatible response with content.
func writeChatResponse(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	response := map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"role": "assistant", "content": content}},
		},
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func TestFromSettingsDisabled(t *testing.T) {
	conf.Reset()
	t.Cleanup(conf.Reset)

	if client := FromSettings(); client != nil {
		t.Fatalf("FromSettings() = %v, want nil", client)
	}
}

func TestFromSettingsRequiresKey(t *testing.T) {
	setAISettings(t, "https://example.com/v1", "model")
	t.Setenv(APIKeyEnv, "")
	t.Setenv(fallbackAPIKeyEnv, "")

	if client := FromSettings(); client != nil {
		t.Fatalf("FromSettings() = %v, want nil without an API key", client)
	}
}

func TestFromSettingsConfigured(t *testing.T) {
	setAISettings(t, "https://example.com/v1/", "model")
	t.Setenv(APIKeyEnv, "secret")
	t.Setenv(fallbackAPIKeyEnv, "")

	client := FromSettings()
	if client == nil {
		t.Fatal("FromSettings() = nil, want a client")
	}
	if client.Endpoint != "https://example.com/v1/" || client.Model != "model" || client.APIKey != "secret" {
		t.Fatalf("unexpected client: %+v", client)
	}
	if client.HTTP == nil || client.HTTP.Timeout <= 0 {
		t.Fatal("client has no timeout")
	}
}

func TestSuggest(t *testing.T) {
	var request chatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat/completions" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer test-key"; got != want {
			t.Errorf("Authorization = %q, want %q", got, want)
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeChatResponse(t, w, codeFence()+"bash\n$ git status\n"+codeFence())
	}))
	defer server.Close()

	client := &Client{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", HTTP: server.Client()}
	suggestion, err := client.Suggest(types.NewCommand("git stauts", "git: 'stauts' is not a git command"))
	if err != nil {
		t.Fatalf("Suggest() error = %v", err)
	}
	if suggestion != "git status" {
		t.Fatalf("Suggest() = %q, want git status", suggestion)
	}
	if request.Model != "test-model" {
		t.Fatalf("request model = %q", request.Model)
	}
	if len(request.Messages) != 2 {
		t.Fatalf("request has %d messages, want 2", len(request.Messages))
	}
	if !strings.Contains(request.Messages[1].Content, "git stauts") || !strings.Contains(request.Messages[1].Content, "not a git command") {
		t.Fatalf("user prompt does not contain the command and output: %q", request.Messages[1].Content)
	}
}

func TestSuggestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	}))
	defer server.Close()

	client := &Client{Endpoint: server.URL, APIKey: "key", Model: "model", HTTP: server.Client()}
	if _, err := client.Suggest(types.NewCommand("git stauts", "out")); err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("Suggest() error = %v, want an HTTP 500 error", err)
	}
}

func TestSuggestInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "not json")
	}))
	defer server.Close()

	client := &Client{Endpoint: server.URL, APIKey: "key", Model: "model", HTTP: server.Client()}
	if _, err := client.Suggest(types.NewCommand("git stauts", "out")); err == nil {
		t.Fatal("Suggest() accepted an invalid JSON response")
	}
}

func TestSuggestNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{\"choices\":[]}")
	}))
	defer server.Close()

	client := &Client{Endpoint: server.URL, APIKey: "key", Model: "model", HTTP: server.Client()}
	if _, err := client.Suggest(types.NewCommand("git stauts", "out")); err == nil {
		t.Fatal("Suggest() accepted a response without choices")
	}
}

func TestSuggestTruncatesOutput(t *testing.T) {
	var request chatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&request)
		writeChatResponse(t, w, "git status")
	}))
	defer server.Close()

	client := &Client{Endpoint: server.URL, APIKey: "key", Model: "model", HTTP: server.Client()}
	output := strings.Repeat("x", maxOutputBytes*2)
	if _, err := client.Suggest(types.NewCommand("git stauts", output)); err != nil {
		t.Fatalf("Suggest() error = %v", err)
	}
	if len(request.Messages[1].Content) > maxOutputBytes+200 {
		t.Fatalf("user prompt is %d bytes, want the output to be truncated", len(request.Messages[1].Content))
	}
	if !strings.Contains(request.Messages[1].Content, "\n\nReturn only the corrected command.") {
		t.Fatalf("unexpected prompt: %q", request.Messages[1].Content)
	}
}

func TestCleanSuggestion(t *testing.T) {
	fenced := codeFence() + "bash\n$ git status\n" + codeFence()
	cases := []struct {
		content string
		want    string
	}{
		{"", ""},
		{"git status", "git status"},
		{"$ git status", "git status"},
		{fenced, "git status"},
		{"explanation\n\ngit status", "explanation"},
	}
	for _, testCase := range cases {
		if got := cleanSuggestion(testCase.content); got != testCase.want {
			t.Errorf("cleanSuggestion(%q) = %q, want %q", testCase.content, got, testCase.want)
		}
	}
}
