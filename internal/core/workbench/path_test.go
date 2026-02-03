package workbench

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name          string
		workbenchName string
		want          string
	}{
		{
			name:          "simple name",
			workbenchName: "auth-backend",
			want:          filepath.Join(home, "wb", "auth-backend"),
		},
		{
			name:          "hyphenated name",
			workbenchName: "my-cool-feature",
			want:          filepath.Join(home, "wb", "my-cool-feature"),
		},
		{
			name:          "single word",
			workbenchName: "frontend",
			want:          filepath.Join(home, "wb", "frontend"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputePath(tt.workbenchName)
			if got != tt.want {
				t.Errorf("ComputePath(%q) = %q, want %q", tt.workbenchName, got, tt.want)
			}
		})
	}
}

func TestDefaultBasePath(t *testing.T) {
	if DefaultBasePath != "wb" {
		t.Errorf("DefaultBasePath = %q, want %q", DefaultBasePath, "wb")
	}
}
