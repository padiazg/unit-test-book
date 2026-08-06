package capturing_stdout_stderr

import (
	"fmt"
	"log"
	"os"
)

type Row struct {
	Label string
	Unit  string
	Value int
}

func PrintTable(rows []Row) {
	if len(rows) == 0 {
		fmt.Println("(empty)")
		return
	}

	fmt.Println("LABEL       | VALUE | UNIT")
	fmt.Println("------------+-------+------")
	for _, r := range rows {
		fmt.Printf("%-10s | %5d | %s\n", r.Label, r.Value, r.Unit)
	}
}

func ReportErrors(errs []error) {
	for _, err := range errs {
		log.Print(err)
	}
}

func WriteDiagnostics(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
