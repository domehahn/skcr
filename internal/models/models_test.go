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

func TestNormalizePlatforms_Dedup(t *testing.T) {
	got, err := NormalizePlatforms([]string{"codex", "codex", "cursor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 unique platforms, got %d: %v", len(got), got)
	}
}

func TestNormalizePlatforms_Invalid(t *testing.T) {
	if _, err := NormalizePlatforms([]string{"codex", "not-a-platform-xyz"}); err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestNormalizePlatforms_All(t *testing.T) {
	got, err := NormalizePlatforms([]string{"all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected multiple platforms from 'all' expansion")
	}
	// Codex must be included.
	found := false
	for _, p := range got {
		if p == "codex" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected codex in 'all' expansion, got %v", got)
	}
}

func TestNormalizePlatforms_AllDedup(t *testing.T) {
	// Passing "all" twice must not produce duplicates.
	got, err := NormalizePlatforms([]string{"all", "all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen := map[string]struct{}{}
	for _, p := range got {
		if _, dup := seen[p]; dup {
			t.Errorf("duplicate platform %q after all+all expansion", p)
		}
		seen[p] = struct{}{}
	}
}

func TestNormalizePlatforms_Empty(t *testing.T) {
	got, err := NormalizePlatforms([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestParsePlatforms_Empty(t *testing.T) {
	got, err := ParsePlatforms("")
	if err != nil || len(got) != 0 {
		t.Errorf("ParsePlatforms('') = %v, %v", got, err)
	}
}
