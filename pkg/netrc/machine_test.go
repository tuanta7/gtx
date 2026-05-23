package netrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMachine_Format(t *testing.T) {
	m := &Machine{
		Name:     "example.com",
		Login:    "user",
		Password: "pass123",
	}

	assert.Equal(t, "machine example.com\nlogin user\npassword pass123\n", string(m.Format()))
}
