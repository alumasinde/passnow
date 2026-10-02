package routes

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestEveryRoutePermissionIsGrantable(t *testing.T) {
	src, err := os.ReadFile("api.go")
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, m := range regexp.MustCompile(`"([a-z_]+\.[a-z_]+(?:\.[a-z_]+)?)"`).FindAllStringSubmatch(string(src), -1) {
		used[m[1]] = true
	}

	files, _ := filepath.Glob("../../migrations/tenant/*.up.sql")
	catalog := map[string]bool{}
	row := regexp.MustCompile(`(?m)^\('([a-z_]+\.[a-z_.]+)','[A-Z]`)
	for _, f := range files {
		b, _ := os.ReadFile(f)
		for _, m := range row.FindAllStringSubmatch(string(b), -1) {
			catalog[m[1]] = true
		}
	}
	if len(catalog) == 0 {
		t.Fatal("no permissions parsed from migrations")
	}

	for code := range used {
		parts := strings.Split(code, ".")
		ok := catalog[code]
		if !ok && len(parts) == 3 { // scoped code: engine only falls back to ".all"
			ok = catalog[parts[0]+"."+parts[1]+".all"]
		}
		if !ok && len(parts) == 2 { // unscoped code: any scoped variant satisfies it
			for c := range catalog {
				if strings.HasPrefix(c, code+".") {
					ok = true
					break
				}
			}
		}
		if !ok {
			t.Errorf("route requires %q but no role can ever be granted it", code)
		}
	}
}