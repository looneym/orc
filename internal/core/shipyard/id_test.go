package shipyard

import "testing"

func TestGenerateShipyardID(t *testing.T) {
	tests := []struct {
		name       string
		currentMax int
		want       string
	}{
		{
			name:       "first shipyard (max=0)",
			currentMax: 0,
			want:       "YARD-001",
		},
		{
			name:       "second shipyard (max=1)",
			currentMax: 1,
			want:       "YARD-002",
		},
		{
			name:       "tenth shipyard (max=9)",
			currentMax: 9,
			want:       "YARD-010",
		},
		{
			name:       "hundredth shipyard (max=99)",
			currentMax: 99,
			want:       "YARD-100",
		},
		{
			name:       "three-digit boundary (max=999)",
			currentMax: 999,
			want:       "YARD-1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateShipyardID(tt.currentMax)
			if got != tt.want {
				t.Errorf("GenerateShipyardID(%d) = %q, want %q", tt.currentMax, got, tt.want)
			}
		})
	}
}

func TestParseShipyardNumber(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{
			name: "valid single digit",
			id:   "YARD-001",
			want: 1,
		},
		{
			name: "valid double digit",
			id:   "YARD-042",
			want: 42,
		},
		{
			name: "valid triple digit",
			id:   "YARD-123",
			want: 123,
		},
		{
			name: "valid four digit",
			id:   "YARD-1000",
			want: 1000,
		},
		{
			name: "invalid format - no dash",
			id:   "YARD001",
			want: -1,
		},
		{
			name: "invalid format - wrong prefix",
			id:   "LIB-001",
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
			got := ParseShipyardNumber(tt.id)
			if got != tt.want {
				t.Errorf("ParseShipyardNumber(%q) = %d, want %d", tt.id, got, tt.want)
			}
		})
	}
}
