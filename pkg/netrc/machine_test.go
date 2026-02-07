package netrc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMachine_Format(t *testing.T) {
	m := &Machine{
		Name:     "example.com",
		Login:    "user",
		Password: "pass123",
	}

	want, _ := os.ReadFile(".netrc")
	got := m.Format()

	assert.Equal(t, want, got)
}
