package value_assertions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "valid positive", input: "42", want: 42},
		{name: "valid zero", input: "0", want: 0},
		{name: "valid negative", input: "-5", want: -5},
		{name: "leading plus", input: "+7", want: 7},
		{name: "empty string", input: "", want: 0},
		{name: "non-numeric", input: "abc", want: 0},
		{name: "with spaces", input: " 3", want: 0},
		{name: "trailing chars", input: "12ab", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseInt(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
