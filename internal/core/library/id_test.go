package library

import "testing"

func TestGenerateLibraryID(t *testing.T) {
	tests := []struct {
		name       string
		currentMax int
		want       string
	}{
		{
			name:       "first library (max=0)",
			currentMax: 0,
			want:       "LIB-001",
		},
		{
			name:       "second library (max=1)",
			currentMax: 1,
			want:       "LIB-002",
		},
		{
			name:       "tenth library (max=9)",
			currentMax: 9,
			want:       "LIB-010",
		},
		{
			name:       "hundredth library (max=99)",
			currentMax: 99,
			want:       "LIB-100",
		},
		{
			name:       "three-digit boundary (max=999)",
			currentMax: 999,
			want:       "LIB-1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateLibraryID(tt.currentMax)
			if got != tt.want {
				t.Errorf("GenerateLibraryID(%d) = %q, want %q", tt.currentMax, got, tt.want)
			}
		})
	}
}

func TestParseLibraryNumber(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{
			name: "valid single digit",
			id:   "LIB-001",
			want: 1,
		},
		{
			name: "valid double digit",
			id:   "LIB-042",
			want: 42,
		},
		{
			name: "valid triple digit",
			id:   "LIB-123",
			want: 123,
		},
		{
			name: "valid four digit",
			id:   "LIB-1000",
			want: 1000,
		},
		{
			name: "invalid format - no dash",
			id:   "LIB001",
			want: -1,
		},
		{
			name: "invalid format - wrong prefix",
			id:   "YARD-001",
			want: -1,
		},
		{
			name: "invalid format - empty",
			id:   "",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLibraryNumber(tt.id)
			if got != tt.want {
				t.Errorf("ParseLibraryNumber(%q) = %d, want %d", tt.id, got, tt.want)
			}
		})
	}
}
