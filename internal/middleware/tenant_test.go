package middleware

import "testing"

func TestFirstPathSegment(t *testing.T) {
	cases := []struct {
		path, slug, rest string
		ok               bool
	}{
		{"/acme/api/v1/visitors", "acme", "/api/v1/visitors", true},
		{"/Acme/", "acme", "/", true},
		{"/api/v1/visitors", "", "/api/v1/visitors", false},
		{"/", "", "/", false},
	}

	for _, tc := range cases {
		slug, rest, ok := firstPathSegment(tc.path)
		if slug != tc.slug || rest != tc.rest || ok != tc.ok {
			t.Fatalf("%q => (%q,%q,%v), want (%q,%q,%v)", tc.path, slug, rest, ok, tc.slug, tc.rest, tc.ok)
		}
	}
}

func TestStripPort(t *testing.T) {
	cases := map[string]string{
		"example.com:8080": "example.com",
		"example.com":      "example.com",
		"":                 "",
	}
	for in, want := range cases {
		if got := stripPort(in); got != want {
			t.Fatalf("stripPort(%q) = %q, want %q", in, got, want)
		}
	}
}
