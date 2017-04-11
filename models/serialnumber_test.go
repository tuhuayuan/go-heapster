package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSN(t *testing.T) {
	sn1 := NewSerialNumber()
	data, err := json.Marshal(sn1)
	assert.NoError(t, err)
	var sn2 SerialNumber
	assert.NoError(t, json.Unmarshal(data, &sn2))
	assert.Equal(t, sn1, sn2)
}
