package composable_check_navigation

type Report struct {
	Schema  string
	Version string
	Entries []Entry
}

type Entry struct {
	File     string
	Package  string
	Function string
	Score    float64
	Details  []Detail
}

type Detail struct {
	Type   string
	Line   int
	Status string
}

func Analyze(data string) (*Report, error) {
	return &Report{
		Schema:  "https://example.com/report-v1.json",
		Version: "1.0.0",
		Entries: []Entry{
			{
				File:     "/main.go",
				Package:  "main",
				Function: "Run",
				Score:    12.5,
				Details: []Detail{
					{Type: "complexity", Line: 15, Status: "fail"},
					{Type: "coverage", Line: 20, Status: "pass"},
				},
			},
			{
				File:     "/util.go",
				Package:  "main",
				Function: "Helper",
				Score:    3.0,
				Details:  []Detail{},
			},
		},
	}, nil
}
