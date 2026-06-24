package output_string_inspection

import (
	"fmt"
	"strings"
)

type ReportRow struct {
	Label string
	Unit  string
	Value int
}

func FormatTable(rows []ReportRow) string {
	if len(rows) == 0 {
		return "(empty)"
	}

	var b strings.Builder
	b.WriteString("LABEL       | VALUE | UNIT\n")
	b.WriteString("------------+-------+------\n")

	for _, r := range rows {
		fmt.Fprintf(&b, "%-10s | %5d | %s\n", r.Label, r.Value, r.Unit)
	}

	return b.String()
}

func FormatCSV(rows []ReportRow) string {
	if len(rows) == 0 {
		return ""
	}

	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "%s,%d,%s\n", r.Label, r.Value, r.Unit)
	}

	return b.String()
}
