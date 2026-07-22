package main

import (
	"testing"

	"byakugan/internal/corpus"
	"byakugan/internal/store"
)

func hitWithRefs(code, section string, refs ...corpus.RelatedSection) store.Hit {
	return store.Hit{
		ID:          code + "-s" + section + "-en-0",
		StatuteCode: code,
		Section:     section,
		Refs:        refs,
	}
}

func TestCollectRefs(t *testing.T) {
	ref := func(code, section string) corpus.RelatedSection {
		return corpus.RelatedSection{Statute: code, Section: section}
	}

	cases := []struct {
		name string
		hits []store.Hit
		max  int
		want []corpus.RelatedSection
	}{
		{
			name: "no refs anywhere",
			hits: []store.Hit{hitWithRefs("DDA1952", "12")},
			max:  3,
			want: nil,
		},
		{
			name: "plain collection in rank order",
			hits: []store.Hit{hitWithRefs("DDA1952", "12", ref("DDA1952", "37"), ref("DDA1952", "39B"))},
			max:  3,
			want: []corpus.RelatedSection{ref("DDA1952", "37"), ref("DDA1952", "39B")},
		},
		{
			name: "ref to an already-retrieved section is skipped",
			hits: []store.Hit{
				hitWithRefs("DDA1952", "12", ref("DDA1952", "37")),
				hitWithRefs("DDA1952", "37"),
			},
			max:  3,
			want: nil,
		},
		{
			name: "same ref in two hits collected once",
			hits: []store.Hit{
				hitWithRefs("AA1960", "37", ref("PC", "304A")),
				hitWithRefs("AA1960", "39", ref("PC", "304A")),
			},
			max:  3,
			want: []corpus.RelatedSection{ref("PC", "304A")},
		},
		{
			name: "cap stops mid-hit, not just between hits",
			hits: []store.Hit{
				hitWithRefs("CPC", "23", ref("CPC", "15"), ref("CPC", "17"), ref("CPC", "62")),
			},
			max:  2,
			want: []corpus.RelatedSection{ref("CPC", "15"), ref("CPC", "17")},
		},
		{
			name: "cap respected across hits",
			hits: []store.Hit{
				hitWithRefs("DDA1952", "12", ref("DDA1952", "37")),
				hitWithRefs("MOA1955", "31", ref("MOA1955", "32"), ref("MOA1955", "14")),
			},
			max:  2,
			want: []corpus.RelatedSection{ref("DDA1952", "37"), ref("MOA1955", "32")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := collectRefs(tc.hits, tc.max)

			if len(got) != len(tc.want) {
				t.Fatalf("got %d refs %v, want %d %v", len(got), got, len(tc.want), tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("refs[%d]: got %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
