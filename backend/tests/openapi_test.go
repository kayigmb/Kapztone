package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

// api/openapi.yaml lists every route with the permission it needs. Read as
// text (no YAML library): a path is a two-space-indented key, its methods
// four-space keys under it, each followed by its x-permission line.
func TestOpenAPIListsEveryRoute(t *testing.T) {
	raw, err := os.ReadFile("../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	spec := map[string]string{} // "METHOD /path" → the method's block
	var path, method string
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, "  /") && strings.HasSuffix(line, ":"):
			path, method = strings.TrimSuffix(strings.TrimSpace(line), ":"), ""
		case path != "" && len(line) > 4 && line[:4] == "    " && line[4] != ' ' && strings.HasSuffix(line, ":"):
			method = strings.ToUpper(strings.TrimSuffix(strings.TrimSpace(line), ":"))
		case line != "" && line[0] != ' ' && line[0] != '#':
			path, method = "", "" // left paths:
		case method != "":
			spec[method+" "+path] += line + "\n"
		}
	}

	for _, rt := range server.Routes(deps(noPing, &auth.Auth{})) {
		name := rt.Method + " " + rt.Path
		block, ok := spec[name]
		switch {
		case !ok:
			t.Errorf("%s is missing from api/openapi.yaml", name)
		case rt.Permission != "" && !strings.Contains(block, "x-permission: "+rt.Permission+"\n"):
			t.Errorf("%s: api/openapi.yaml should say x-permission: %s", name, rt.Permission)
		case rt.Access == server.Public && !strings.Contains(block, "security: []"):
			t.Errorf("%s is public: api/openapi.yaml should say security: []", name)
		}
		delete(spec, name)
	}
	for name := range spec {
		t.Errorf("api/openapi.yaml lists %s, which is not a route", name)
	}
}
