package models

import "testing"

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{" GitLab ", "gitlab-duo", false},
		{"github-copilot-chat", "github-copilot", false},
		{"open-hands", "openhands", false},
		{"codex", "codex", false},
		{"antigravity", "antigravity", false},
		{"Amazon Q Developer", "amazon-q", false},
		{"kilo-code", "kilocode", false},
		{"qwen-code", "qwen", false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		got, err := NormalizePlatform(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("NormalizePlatform(%q): expected error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizePlatform(%q) unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizePlatform(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParsePlatforms(t *testing.T) {
	got, err := ParsePlatforms("gitlab, codex, gitlab, github")
	if err != nil {
		t.Fatalf("ParsePlatforms unexpected error: %v", err)
	}
	want := []string{"gitlab-duo", "codex", "github-copilot"}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}

	empty, err := ParsePlatforms("   ")
	if err != nil || len(empty) != 0 {
		t.Fatalf("expected empty result, got=%v err=%v", empty, err)
	}

	if _, err := ParsePlatforms("codex,invalid"); err == nil {
		t.Fatal("expected error for invalid platform")
	}
}

func TestAllConcretePlatformsExcludesAllAndIncludesExtendedTools(t *testing.T) {
	got := AllConcretePlatforms()
	seen := map[string]struct{}{}
	for _, platform := range got {
		if platform == "all" {
			t.Fatal("AllConcretePlatforms must not include all")
		}
		if _, ok := seen[platform]; ok {
			t.Fatalf("AllConcretePlatforms contains duplicate %q: %#v", platform, got)
		}
		seen[platform] = struct{}{}
	}
	for _, want := range []string{"codex", "generic", "antigravity", "amazon-q", "qwen"} {
		found := false
		for _, platform := range got {
			if platform == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("AllConcretePlatforms missing %q: %#v", want, got)
		}
	}
}
