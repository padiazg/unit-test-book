package output_string_inspection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkFormatTableFn func(*testing.T, string)

var checkFormatTable = func(fns ...checkFormatTableFn) []checkFormatTableFn { return fns }

func TestFormatTable(t *testing.T) {
	checkContains := func(want string) checkFormatTableFn {
		return func(t *testing.T, got string) {
			t.Helper()
			assert.Contains(t, got, want)
		}
	}

	checkNotContains := func(want string) checkFormatTableFn {
		return func(t *testing.T, got string) {
			t.Helper()
			assert.NotContains(t, got, want)
		}
	}

	tests := []struct {
		name   string
		rows   []ReportRow
		checks []checkFormatTableFn
	}{
		{
			name:   "empty rows",
			rows:   nil,
			checks: checkFormatTable(
				checkContains("(empty)"),
				checkNotContains("LABEL"),
			),
		},
		{
			name: "single row",
			rows: []ReportRow{{Label: "CPU", Value: 45, Unit: "%"}},
			checks: checkFormatTable(
				checkContains("CPU"),
				checkContains("45"),
				checkContains("%"),
				checkContains("LABEL"),
				checkNotContains("MEM"),
			),
		},
		{
			name: "multiple rows",
			rows: []ReportRow{
				{Label: "CPU", Value: 45, Unit: "%"},
				{Label: "MEM", Value: 2048, Unit: "MB"},
				{Label: "DISK", Value: 500, Unit: "GB"},
			},
			checks: checkFormatTable(
				checkContains("CPU"),
				checkContains("MEM"),
				checkContains("DISK"),
				checkContains("45"),
				checkContains("2048"),
				checkContains("500"),
				checkNotContains("(empty)"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTable(tt.rows)
			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}

func TestFormatCSV(t *testing.T) {
	tests := []struct {
		name string
		rows []ReportRow
		want string
	}{
		{name: "empty", rows: nil, want: ""},
		{name: "single", rows: []ReportRow{{Label: "CPU", Value: 45, Unit: "%"}}, want: "CPU,45,%\n"},
		{name: "multiple", rows: []ReportRow{
			{Label: "CPU", Value: 45, Unit: "%"},
			{Label: "MEM", Value: 2048, Unit: "MB"},
		}, want: "CPU,45,%\nMEM,2048,MB\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCSV(tt.rows)
			assert.Equal(t, tt.want, got)
		})
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	for {
		idx := indexOf(s, '\n')
		if idx < 0 {
			lines = append(lines, s)
			break
		}
		lines = append(lines, s[:idx])
		s = s[idx+1:]
	}
	return lines
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
