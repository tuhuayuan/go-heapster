package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelSet(t *testing.T) {
	const (
		testLabelName     = "__test_name"
		testLabelPassword = "__test_passwored"
	)
	type Metric struct {
		Labels LabelSet `json:"lablels"`
	}

	m1 := Metric{
		Labels: LabelSet{
			testLabelName:     "tuhuayuan",
			testLabelPassword: "123456",
		},
	}
	fmt.Println(m1)
	data, err := json.Marshal(m1.Labels)
	assert.NoError(t, err)
	fmt.Println(string(data))
}
