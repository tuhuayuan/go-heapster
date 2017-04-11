package models

import "testing"
import "github.com/stretchr/testify/assert"
import "time"

func TestDuration(t *testing.T) {
	d1, err := ParseDuration("3s")
	assert.NoError(t, err)
	assert.EqualValues(t, 3*time.Second, d1)
}
