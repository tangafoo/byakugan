package store

import "testing"

func TestFilterByDistance(t *testing.T) {
	hits := func(distances ...float64) []Hit {
		out := make([]Hit, len(distances))
		for i, d := range distances {
			out[i] = Hit{ID: "h", Distance: d}
		}
		return out
	}

	cases := []struct {
		name string
		in   []Hit
		max  float64
		want int
	}{
		{"zero means off", hits(0.2, 0.9, 1.5), 0, 3},
		{"negative means off", hits(0.9), -1, 1},
		{"below threshold kept", hits(0.5, 0.6), 0.71, 2},
		{"above threshold dropped", hits(0.72, 0.9), 0.71, 0},
		{"exactly at threshold kept", hits(0.71), 0.71, 1},
		{"mixed", hits(0.3, 0.71, 0.7100001, 1.2), 0.71, 2},
		{"all filtered leaves empty not nil-panic", hits(0.9, 0.95), 0.5, 0},
		{"empty input survives", nil, 0.71, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterByDistance(tc.in, tc.max)
			if len(got) != tc.want {
				t.Errorf("kept %d hits, want %d", len(got), tc.want)
			}
			for _, h := range got {
				if tc.max > 0 && h.Distance > tc.max {
					t.Errorf("hit with distance %v survived threshold %v", h.Distance, tc.max)
				}
			}
		})
	}
}
