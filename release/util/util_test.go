package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extraTags(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		existing []string
		want     []string
		wantErr  bool
	}{
		// Basic behavior
		{name: "New Major",
			target:   "v2.0.0",
			existing: []string{"v1.0.0", "v1.0.1", "v1.1.0"},
			want:     []string{"v2", "v2.0", "latest"},
		},
		{name: "New Minor",
			target:   "v1.2.0",
			existing: []string{"v1.0.0", "v1.0.1", "v1.1.0"},
			want:     []string{"v1", "v1.2", "latest"},
		},
		{name: "New Patch",
			target:   "v1.1.1",
			existing: []string{"v1.0.0", "v1.0.1", "v1.1.0"},
			want:     []string{"v1", "v1.1", "latest"},
		},
		// First Release (no existing tags)
		{name: "First Release - Major",
			target:   "v1.0.0",
			existing: []string{},
			want:     []string{"v1", "v1.0", "latest"},
		},
		{name: "First Release - Minor",
			target:   "v0.1.0",
			existing: []string{},
			want:     []string{"v0", "v0.1", "latest"},
		},
		{name: "First Release - Patch",
			target:   "v0.0.1",
			existing: []string{},
			want:     []string{"v0", "v0.0", "latest"},
		},
		// Not Latest
		{name: "Patch Older - Greator Minor",
			target:   "v1.1.1",
			existing: []string{"v2.0.0", "v2.1.0", "v2.2.0", "v1.2.0", "v1.1.0"},
			want:     []string{"v1.1"},
		},
		{name: "Patch Older - Latest Minor",
			target:   "v1.2.1",
			existing: []string{"v2.0.0", "v2.1.0", "v2.2.0", "v1.2.0", "v1.1.0"},
			want:     []string{"v1", "v1.2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtraTags(tt.target, tt.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("extraTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.ElementsMatch(t, got, tt.want) {
				t.Errorf("unexpected extra tags, got = %v, want = %v", got, tt.want)
			}
		})
	}
}
