package classic_table_driven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerson(t *testing.T) {
	tests := []struct {
		name    string
		person  string
		email   string
		age     int
		wantErr bool
	}{
		{
			name:    "valid person",
			person:  "Alice",
			email:   "alice@example.com",
			age:     30,
			wantErr: false,
		},
		{
			name:    "empty name",
			person:  "",
			email:   "bob@example.com",
			age:     25,
			wantErr: true,
		},
		{
			name:    "whitespace name",
			person:  "  ",
			email:   "bob@example.com",
			age:     25,
			wantErr: true,
		},
		{
			name:    "missing email",
			person:  "Bob",
			email:   "",
			age:     25,
			wantErr: true,
		},
		{
			name:    "invalid email format",
			person:  "Bob",
			email:   "not-an-email",
			age:     25,
			wantErr: true,
		},
		{
			name:    "negative age",
			person:  "Charlie",
			email:   "charlie@example.com",
			age:     -1,
			wantErr: true,
		},
		{
			name:    "age too high",
			person:  "Diana",
			email:   "diana@example.com",
			age:     200,
			wantErr: true,
		},
		{
			name:    "age zero",
			person:  "Eve",
			email:   "eve@example.com",
			age:     0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPerson(tt.person, tt.email, tt.age)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, p)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, p)
			assert.Equal(t, tt.person, p.Name)
			assert.Equal(t, tt.email, p.Email)
			assert.Equal(t, tt.age, p.Age)
		})
	}
}
