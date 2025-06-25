package shared

import "testing"

func TestExtractChapterRangeFromTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[One Pace][8-11] adas", "8-11"},
		{"[One Pace][42] single1", "42-42"},
		{"[One Pace][3, 153-156] single2", "3-3"},
		{"[One Pace][123-124, 520] tail", "123-124"},
		{"nothingatall", ""},
	}

	for _, tc := range tests {
		got := ExtractChapterRangeFromTitle(tc.input)
		if got != tc.expected {
			t.Errorf("input %q: got %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestRoughExtractChapterFromTitle(t *testing.T) {
	tests := []struct {
		input       string
		want        string
		expectRange bool
	}{
		{"Chapter 01", "01", false},
		{"Chapter 22", "22", false},
		{"Chapter 583240", "583240", false},
		{"Chapter33", "33", false},
		{"Chapter 22  55", "22", false},
		{"Chapter", "00", false},
		{"NotAChapter", "00", false},
		{"Chapter 841-845", "841-845", true},
		{"Chapter35-36", "35-36", true},
		{"Chapters 35-36", "35-36", true},
		{"One Pace] Paced One Piece - Thriller Bark Episode 18 [720p][2295F0A1].mkv", "18", false},
		{"[One Pace] Chapter 831-832 [720p][DF6B6FEC].mkv", "831-832", true},
	}

	for _, tt := range tests {
		got, isRange := RoughExtractChapterFromTitle(tt.input)
		if got != tt.want || isRange != tt.expectRange {
			t.Errorf("input %q: got (%q, %v), want (%q, %v)", tt.input, got, isRange, tt.want, tt.expectRange)
		}
	}
}
