package output_string_inspection

import (
	"fmt"
	"strings"
)

type ReportRow struct {
	Label string
	Value int
	Unit  string
}

func FormatTable(rows []ReportRow) string {
	if len(rows) == 0 {
		return "(empty)"
	}

	var b strings.Builder
	b.WriteString("LABEL       | VALUE | UNIT\n")
	b.WriteString("------------+-------+------\n")

	for _, r := range rows {
		b.WriteString(fmt.Sprintf("%-10s | %5d | %s\n", r.Label, r.Value, r.Unit))
	}

	return b.String()
}

func FormatCSV(rows []ReportRow) string {
	if len(rows) == 0 {
		return ""
	}

	var b strings.Builder
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("%s,%d,%s\n", r.Label, r.Value, r.Unit))
	}

	return b.String()
}
