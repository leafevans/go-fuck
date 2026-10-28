package main

import (
	"strings"
	"testing"
)

// TestRewrite checks that rewrite replaces the marked block and only it.
func TestRewrite(t *testing.T) {
	tests := []struct {
		name     string
		document string
		names    []string
		want     string
	}{
		{
			name:     "replaces an existing list",
			document: "# Doc\n\n<!-- RULES_LIST_START -->\n- `old`\n<!-- RULES_LIST_END -->\n\nTail.\n",
			names:    []string{"alpha", "beta"},
			want:     "# Doc\n\n<!-- RULES_LIST_START -->\n- `alpha`\n- `beta`\n<!-- RULES_LIST_END -->\n\nTail.\n",
		},
		{
			name:     "empty list",
			document: "<!-- RULES_LIST_START -->\nignored\n<!-- RULES_LIST_END -->\n",
			names:    nil,
			want:     "<!-- RULES_LIST_START -->\n<!-- RULES_LIST_END -->\n",
		},
		{
			name:     "preserves crlf line endings",
			document: "# Doc\r\n\r\n<!-- RULES_LIST_START -->\r\n- `old`\r\n<!-- RULES_LIST_END -->\r\n",
			names:    []string{"alpha"},
			want:     "# Doc\r\n\r\n<!-- RULES_LIST_START -->\r\n- `alpha`\r\n<!-- RULES_LIST_END -->\r\n",
		},
		{
			name:     "keeps text after the end marker",
			document: "a\n<!-- RULES_LIST_START -->\nb\n<!-- RULES_LIST_END -->\nc\n",
			names:    []string{"alpha"},
			want:     "a\n<!-- RULES_LIST_START -->\n- `alpha`\n<!-- RULES_LIST_END -->\nc\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := rewrite(test.document, test.names)
			if err != nil {
				t.Fatalf("rewrite() error = %v", err)
			}
			if got != test.want {
				t.Errorf("rewrite() = %q, want %q", got, test.want)
			}
		})
	}
}

// TestRewriteErrors checks that malformed documents are reported.
func TestRewriteErrors(t *testing.T) {
	tests := []struct {
		name     string
		document string
		wantErr  string
	}{
		{
			name:     "missing start marker",
			document: "no markers here",
			wantErr:  "missing",
		},
		{
			name:     "missing end marker",
			document: "<!-- RULES_LIST_START -->\n- `old`\n",
			wantErr:  "missing",
		},
		{
			name:     "duplicate start marker",
			document: "<!-- RULES_LIST_START -->\n<!-- RULES_LIST_START -->\n<!-- RULES_LIST_END -->\n",
			wantErr:  "more than one",
		},
		{
			name:     "end marker before start marker",
			document: "<!-- RULES_LIST_END -->\n<!-- RULES_LIST_START -->\n",
			wantErr:  "before",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := rewrite(test.document, []string{"alpha"})
			if err == nil {
				t.Fatal("rewrite() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("rewrite() error = %q, want it to contain %q", err, test.wantErr)
			}
		})
	}
}
