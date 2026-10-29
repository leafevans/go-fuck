package entrypoints

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leafevans/go-fuck/internal/ai"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/types"
)

func TestAIFallbackDisabled(t *testing.T) {
	resetEntrypointsState(t)

	if got := aiFallback(types.NewCommand("git stauts", "error")); got != nil {
		t.Fatalf("aiFallback() = %v, want nil when the fallback is disabled", got)
	}
}

func TestAIFallbackSuggestion(t *testing.T) {
	resetEntrypointsState(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "git status"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv(ai.APIKeyEnv, "test-key")
	conf.Set(&conf.Settings{AIEnabled: true, AIEndpoint: server.URL, AIModel: "model", AITimeout: 5})
	t.Cleanup(conf.Reset)

	got := aiFallback(types.NewCommand("git stauts", "is not a git command"))
	if len(got) != 1 || got[0].Script != "git status" {
		t.Fatalf("aiFallback() = %v, want the AI suggestion", got)
	}
}

func TestAIFallbackFailure(t *testing.T) {
	resetEntrypointsState(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv(ai.APIKeyEnv, "test-key")
	conf.Set(&conf.Settings{AIEnabled: true, AIEndpoint: server.URL, AIModel: "model", AITimeout: 5})
	t.Cleanup(conf.Reset)

	if got := aiFallback(types.NewCommand("git stauts", "error")); got != nil {
		t.Fatalf("aiFallback() = %v, want nil on failure", got)
	}
}
