package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "Simple common prefix",
			input:    []string{"prefix-file1.txt", "prefix-file2.txt", "prefix-file3.txt"},
			expected: "prefix-",
		},
		{
			name:     "Chinese prefix with separator",
			input:    []string{"【歌单】song1.mp3", "【歌单】song2.mp3", "【歌单】song3.mp3"},
			expected: "【歌单】",
		},
		{
			name:     "No common prefix",
			input:    []string{"abc.txt", "def.txt", "ghi.txt"},
			expected: "",
		},
		{
			name:     "Single file",
			input:    []string{"file.txt"},
			expected: "",
		},
		{
			name:     "Empty list",
			input:    []string{},
			expected: "",
		},
		{
			name:     "Prefix with space separator",
			input:    []string{"Common Prefix file1.mp3", "Common Prefix file2.mp3"},
			expected: "Common Prefix ",
		},
		{
			name: "Complex Chinese prefix",
			input: []string{
				"【Music Collection】Taylor Swift Greatest Hits p01 Shake It Off.m4a",
				"【Music Collection】Taylor Swift Greatest Hits p02 Blank Space.m4a",
				"【Music Collection】Taylor Swift Greatest Hits p03 Love Story.m4a",
			},
			expected: "【Music Collection】Taylor Swift Greatest Hits ",
		},
		{
			name: "Files with numbers and Chinese prefix",
			input: []string{
				"01-【Playlist】Song Name 1.mp3",
				"02-【Playlist】Song Name 2.mp3",
				"03-【Playlist】Song Name 3.mp3",
			},
			expected: "0",
		},
		{
			name:     "Prefix too short (less than 3 chars)",
			input:    []string{"a1.txt", "a2.txt", "a3.txt"},
			expected: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCommonPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("findCommonPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFindMajorityPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "Majority with outliers",
			input: []string{
				"Artist-Song1.mp3",
				"Artist-Song2.mp3",
				"Artist-Song3.mp3",
				"Artist-Song4.mp3",
				"Artist-Song5.mp3",
				"OtherArtist&Artist-Collab1.mp3",
				"AnotherArtist&Artist-Collab2.mp3",
			},
			expected: "Artist-",
		},
		{
			name: "70% threshold",
			input: []string{
				"[Playlist] Song1.mp3",
				"[Playlist] Song2.mp3",
				"[Playlist] Song3.mp3",
				"[Playlist] Song4.mp3",
				"[Playlist] Song5.mp3",
				"[Playlist] Song6.mp3",
				"[Playlist] Song7.mp3",
				"Other-Song8.mp3",
				"Other-Song9.mp3",
				"Other-Song10.mp3",
			},
			expected: "[Playlist] ",
		},
		{
			name: "No majority",
			input: []string{
				"A-file1.txt",
				"B-file2.txt",
				"C-file3.txt",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMajorityPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("findMajorityPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCollectFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test structure
	dirs := []string{
		filepath.Join(tmpDir, "music"),
		filepath.Join(tmpDir, "music", "@eaDir"),
		filepath.Join(tmpDir, "music", "subdir"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create test files
	files := []struct {
		path string
		dir  string
	}{
		{filepath.Join(tmpDir, "music", "song1.mp3"), "music"},
		{filepath.Join(tmpDir, "music", "song2.m4a"), "music"},
		{filepath.Join(tmpDir, "music", "song3.flac"), "music"},
		{filepath.Join(tmpDir, "music", "readme.txt"), "music"},
		{filepath.Join(tmpDir, "music", "@eaDir", "thumb.jpg"), "@eaDir"},
		{filepath.Join(tmpDir, "music", "subdir", "song4.mp3"), "subdir"},
	}
	for _, f := range files {
		if err := os.WriteFile(f.path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name          string
		excludeDirs   []string
		extensions    []string
		expectedMin   int
		expectedMax   int
		shouldExclude string
	}{
		{
			name:        "All files, no exclusions",
			excludeDirs: []string{},
			extensions:  []string{},
			expectedMin: 6,
			expectedMax: 6,
		},
		{
			name:          "Exclude @eaDir",
			excludeDirs:   []string{"@eaDir"},
			extensions:    []string{},
			expectedMin:   5,
			expectedMax:   5,
			shouldExclude: "@eaDir",
		},
		{
			name:        "Only audio files (mp3, m4a, flac)",
			excludeDirs: []string{},
			extensions:  []string{"mp3", "m4a", "flac"},
			expectedMin: 4,
			expectedMax: 4,
		},
		{
			name:        "Audio files excluding @eaDir",
			excludeDirs: []string{"@eaDir"},
			extensions:  []string{"mp3", "m4a", "flac"},
			expectedMin: 4,
			expectedMax: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := collectFiles(tmpDir, tt.excludeDirs, tt.extensions)
			if err != nil {
				t.Errorf("collectFiles() error = %v", err)
				return
			}

			if len(result) < tt.expectedMin || len(result) > tt.expectedMax {
				t.Errorf("collectFiles() got %d files, want between %d and %d",
					len(result), tt.expectedMin, tt.expectedMax)
			}

			// Check that excluded directory is not present
			if tt.shouldExclude != "" {
				for _, file := range result {
					if filepath.Base(filepath.Dir(file)) == tt.shouldExclude {
						t.Errorf("Found file in excluded directory: %s", file)
					}
				}
			}
		})
	}
}

func TestGroupFilesByDirectory(t *testing.T) {
	files := []string{
		"/music/dir1/song1.mp3",
		"/music/dir1/song2.mp3",
		"/music/dir2/song3.mp3",
		"/music/dir3/song4.mp3",
		"/music/dir3/song5.mp3",
		"/music/dir3/song6.mp3",
	}

	result := groupFilesByDirectory(files)

	expectedGroups := 3
	if len(result) != expectedGroups {
		t.Errorf("groupFilesByDirectory() got %d groups, want %d", len(result), expectedGroups)
	}

	if len(result["/music/dir1"]) != 2 {
		t.Errorf("dir1 should have 2 files, got %d", len(result["/music/dir1"]))
	}

	if len(result["/music/dir2"]) != 1 {
		t.Errorf("dir2 should have 1 file, got %d", len(result["/music/dir2"]))
	}

	if len(result["/music/dir3"]) != 3 {
		t.Errorf("dir3 should have 3 files, got %d", len(result["/music/dir3"]))
	}
}

func BenchmarkFindCommonPrefix(b *testing.B) {
	files := []string{
		"01-【Music Collection】Taylor Swift Greatest Hits p01 Shake It Off.m4a",
		"02-【Music Collection】Taylor Swift Greatest Hits p02 Blank Space.m4a",
		"03-【Music Collection】Taylor Swift Greatest Hits p03 Love Story.m4a",
		"04-【Music Collection】Taylor Swift Greatest Hits p04 You Belong With Me.m4a",
		"05-【Music Collection】Taylor Swift Greatest Hits p05 We Are Never Getting Back Together.m4a",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findCommonPrefix(files)
	}
}
