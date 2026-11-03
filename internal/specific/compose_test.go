package specific

import (
	"strings"
	"testing"
)

// composeFixture is a small Compose file used by the parser tests.
const composeFixture = `# example stack
version: "3.9"
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
  worker:
    build: .
    depends_on:
      - db
volumes:
  data:
`

func TestParseComposeServices(t *testing.T) {
	services := parseComposeServices(composeFixture)
	want := []string{"db", "web", "worker"}
	if strings.Join(services, ",") != strings.Join(want, ",") {
		t.Fatalf("parseComposeServices() = %q, want %q", services, want)
	}
}

func TestParseComposeServicesEmpty(t *testing.T) {
	if services := parseComposeServices(""); len(services) != 0 {
		t.Fatalf("parseComposeServices() = %q, want no services", services)
	}
	if services := parseComposeServices("version: \"3\"\n"); len(services) != 0 {
		t.Fatalf("parseComposeServices() = %q, want no services", services)
	}
}
