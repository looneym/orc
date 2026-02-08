package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCompareDirs tests the compareDirs function that compares skill directories
func TestCompareDirs(t *testing.T) {
	tests := []struct {
		name        string
		setupSrc    func(dir string) // Setup source directory
		setupDst    func(dir string) // Setup destination directory
		wantMissing []string
		wantStale   []string
	}{
		{
			name: "all present and matching",
			setupSrc: func(dir string) {
				// Create skill1 with file
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("content1"), 0644)
				// Create skill2 with file
				skill2 := filepath.Join(dir, "skill2")
				os.MkdirAll(skill2, 0755)
				os.WriteFile(filepath.Join(skill2, "SKILL.md"), []byte("content2"), 0644)
			},
			setupDst: func(dir string) {
				// Same content
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("content1"), 0644)
				skill2 := filepath.Join(dir, "skill2")
				os.MkdirAll(skill2, 0755)
				os.WriteFile(filepath.Join(skill2, "SKILL.md"), []byte("content2"), 0644)
			},
			wantMissing: nil,
			wantStale:   nil,
		},
		{
			name: "missing skill directory",
			setupSrc: func(dir string) {
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("content1"), 0644)
				skill2 := filepath.Join(dir, "skill2")
				os.MkdirAll(skill2, 0755)
				os.WriteFile(filepath.Join(skill2, "SKILL.md"), []byte("content2"), 0644)
			},
			setupDst: func(dir string) {
				// Only skill1 exists
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("content1"), 0644)
			},
			wantMissing: []string{"skill2"},
			wantStale:   nil,
		},
		{
			name: "stale skill with different content",
			setupSrc: func(dir string) {
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("new content"), 0644)
			},
			setupDst: func(dir string) {
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
				os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("old content"), 0644)
			},
			wantMissing: nil,
			wantStale:   []string{"skill1"},
		},
		{
			name: "source directory does not exist",
			setupSrc: func(dir string) {
				// Don't create anything - we'll use a non-existent path
			},
			setupDst: func(dir string) {
				skill1 := filepath.Join(dir, "skill1")
				os.MkdirAll(skill1, 0755)
			},
			wantMissing: nil,
			wantStale:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directories
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			// Setup
			tt.setupSrc(srcDir)
			tt.setupDst(dstDir)

			// For "source does not exist" test, use non-existent path
			srcPath := srcDir
			if tt.name == "source directory does not exist" {
				srcPath = filepath.Join(srcDir, "nonexistent")
			}

			// Run
			gotMissing, gotStale := compareDirs(srcPath, dstDir)

			// Verify missing
			if !stringSliceEqual(gotMissing, tt.wantMissing) {
				t.Errorf("missing = %v, want %v", gotMissing, tt.wantMissing)
			}

			// Verify stale
			if !stringSliceEqual(gotStale, tt.wantStale) {
				t.Errorf("stale = %v, want %v", gotStale, tt.wantStale)
			}
		})
	}
}

// TestCompareFiles tests the compareFiles function for hooks/tmux
func TestCompareFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupSrc    func(dir string)
		setupDst    func(dir string)
		wantMissing []string
		wantStale   []string
	}{
		{
			name: "all files present and matching",
			setupSrc: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("#!/bin/bash\necho hello"), 0644)
				os.WriteFile(filepath.Join(dir, "hook2.sh"), []byte("#!/bin/bash\necho world"), 0644)
			},
			setupDst: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("#!/bin/bash\necho hello"), 0644)
				os.WriteFile(filepath.Join(dir, "hook2.sh"), []byte("#!/bin/bash\necho world"), 0644)
			},
			wantMissing: nil,
			wantStale:   nil,
		},
		{
			name: "missing file",
			setupSrc: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("content1"), 0644)
				os.WriteFile(filepath.Join(dir, "hook2.sh"), []byte("content2"), 0644)
			},
			setupDst: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("content1"), 0644)
				// hook2.sh missing
			},
			wantMissing: []string{"hook2.sh"},
			wantStale:   nil,
		},
		{
			name: "stale file with different content",
			setupSrc: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("new version"), 0644)
			},
			setupDst: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("old version"), 0644)
			},
			wantMissing: nil,
			wantStale:   []string{"hook1.sh"},
		},
		{
			name: "source does not exist",
			setupSrc: func(dir string) {
				// Don't create - use nonexistent path
			},
			setupDst: func(dir string) {
				os.WriteFile(filepath.Join(dir, "hook1.sh"), []byte("content"), 0644)
			},
			wantMissing: nil,
			wantStale:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			tt.setupSrc(srcDir)
			tt.setupDst(dstDir)

			srcPath := srcDir
			if tt.name == "source does not exist" {
				srcPath = filepath.Join(srcDir, "nonexistent")
			}

			gotMissing, gotStale := compareFiles(srcPath, dstDir)

			if !stringSliceEqual(gotMissing, tt.wantMissing) {
				t.Errorf("missing = %v, want %v", gotMissing, tt.wantMissing)
			}
			if !stringSliceEqual(gotStale, tt.wantStale) {
				t.Errorf("stale = %v, want %v", gotStale, tt.wantStale)
			}
		})
	}
}

// TestDirsEqual tests the dirsEqual helper function
func TestDirsEqual(t *testing.T) {
	tests := []struct {
		name      string
		setupDir1 func(dir string)
		setupDir2 func(dir string)
		want      bool
	}{
		{
			name: "identical directories",
			setupDir1: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
			},
			setupDir2: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
			},
			want: true,
		},
		{
			name: "different file count",
			setupDir1: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
			},
			setupDir2: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
				// file2.txt missing
			},
			want: false,
		},
		{
			name: "different content",
			setupDir1: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("new content"), 0644)
			},
			setupDir2: func(dir string) {
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("old content"), 0644)
			},
			want: false,
		},
		{
			name: "nested directories identical",
			setupDir1: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("nested content"), 0644)
			},
			setupDir2: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("nested content"), 0644)
			},
			want: true,
		},
		{
			name: "nested directories different",
			setupDir1: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("new nested"), 0644)
			},
			setupDir2: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("old nested"), 0644)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir1 := t.TempDir()
			dir2 := t.TempDir()

			tt.setupDir1(dir1)
			tt.setupDir2(dir2)

			got := dirsEqual(dir1, dir2)
			if got != tt.want {
				t.Errorf("dirsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckResult tests the CheckResult struct behavior
func TestCheckResult(t *testing.T) {
	tests := []struct {
		name   string
		result CheckResult
	}{
		{
			name:   "passing check",
			result: CheckResult{Name: "Test", Status: "✓"},
		},
		{
			name:   "warning check",
			result: CheckResult{Name: "Test", Status: "⚠", Details: "  Some warning"},
		},
		{
			name:   "failing check",
			result: CheckResult{Name: "Test", Status: "✗", Details: "  Error details"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Name == "" {
				t.Error("Name should not be empty")
			}
			if tt.result.Status == "" {
				t.Error("Status should not be empty")
			}
		})
	}
}

// stringSliceEqual compares two string slices
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
