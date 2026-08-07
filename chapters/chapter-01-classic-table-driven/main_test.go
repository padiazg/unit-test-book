package classic_table_driven

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntries_ThresholdExceeded(t *testing.T) {
	const fn = "ThresholdExceeded"
	tests := []struct {
		name      string
		entries   *Entries
		threshold float64
		want      bool
	}{
		{
			name:      "empty_list_returns_false",
			entries:   &Entries{List: []CRAPEntry{}},
			threshold: 100.0,
			want:      false,
		},
		{
			name: "all_below_threshold",
			entries: &Entries{List: []CRAPEntry{
				{FuncName: "a", EffectiveCRAP: 50},
				{FuncName: "b", EffectiveCRAP: 30},
			}},
			threshold: 100.0,
			want:      false,
		},
		{
			name: "all_at_threshold_not_exceeded",
			entries: &Entries{List: []CRAPEntry{
				{FuncName: "exact", EffectiveCRAP: 100},
			}},
			threshold: 100.0,
			want:      false,
		},
		{
			name: "one_entry_exceeding_threshold",
			entries: &Entries{List: []CRAPEntry{
				{FuncName: "low", EffectiveCRAP: 30},
				{FuncName: "high", EffectiveCRAP: 200},
			}},
			threshold: 100.0,
			want:      true,
		},
		{
			name: "all_above_threshold",
			entries: &Entries{List: []CRAPEntry{
				{FuncName: "a", EffectiveCRAP: 200},
				{FuncName: "b", EffectiveCRAP: 150},
			}},
			threshold: 100.0,
			want:      true,
		},
		{
			name: "single_entry_below_threshold",
			entries: &Entries{List: []CRAPEntry{
				{FuncName: "single", EffectiveCRAP: 50},
			}},
			threshold: 100.0,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entries.ThresholdExceeded(tt.threshold)
			assert.Equal(t, tt.want, got, "%s: %s", fn, tt.name)
		})
	}
}
