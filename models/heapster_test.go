package models

import (
	"context"
	"fmt"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func TestApplyNotifiers(t *testing.T) {
	config := `{
        "notifiers": {
            "sms": {
                "type":            "unicom",
				"unicom_sp":       "103905",
				"unicom_username": "zz_sj",
				"unicom_password": "www.zonst.org",
				"targets":         ["13879156403", "15507911970"]
            }
        }
    }`
	hp1 := Heapster{}
	assert.NoError(t, json.Unmarshal([]byte(config), &hp1))
	fmt.Println(hp1)
	hp := &Heapster{
		Notifiers: map[string]interface{}{
			"sms": map[string]interface{}{
				"type":     "unicom",
				"sp":       "103905",
				"username": "zz_sj",
				"password": "www.zonst.org",
				"targets":  []string{"13879156403", "15507911970"},
			},
		},
	}
	notifiers, err := hp.GetApplyNotifiers(context.Background())
	assert.NoError(t, err)
	fmt.Println(notifiers)
}
