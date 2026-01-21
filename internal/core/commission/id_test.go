package commission

import "testing"

func TestGenerateCommissionID(t *testing.T) {
	tests := []struct {
		name       string
		currentMax int
		want       string
	}{
		{
			name:       "first commission (max=0)",
			currentMax: 0,
			want:       "COMM-001",
		},
		{
			name:       "second commission (max=1)",
			currentMax: 1,
			want:       "COMM-002",
		},
		{
			name:       "tenth commission (max=9)",
			currentMax: 9,
			want:       "COMM-010",
		},
		{
			name:       "hundredth commission (max=99)",
			currentMax: 99,
			want:       "COMM-100",
		},
		{
			name:       "three-digit boundary (max=999)",
			currentMax: 999,
			want:       "COMM-1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCommissionID(tt.currentMax)
			if got != tt.want {
				t.Errorf("GenerateCommissionID(%d) = %q, want %q", tt.currentMax, got, tt.want)
			}
		})
	}
}

func TestParseCommissionNumber(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{
			name: "valid single digit",
			id:   "COMM-001",
			want: 1,
		},
		{
			name: "valid double digit",
			id:   "COMM-042",
			want: 42,
		},
		{
			name: "valid triple digit",
			id:   "COMM-123",
			want: 123,
		},
		{
			name: "valid four digit",
			id:   "COMM-1000",
			want: 1000,
		},
		{
			name: "invalid format - no dash",
			id:   "COMM001",
			want: -1,
		},
		{
			name: "invalid format - wrong prefix",
			id:   "GROVE-001",
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
			got := ParseCommissionNumber(tt.id)
			if got != tt.want {
				t.Errorf("ParseCommissionNumber(%q) = %d, want %d", tt.id, got, tt.want)
			}
		})
	}
}
