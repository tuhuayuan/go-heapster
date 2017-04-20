package daemons

import (
	"testing"
	"time"
	"zonst/qipai/logagent/utils"

	"github.com/stretchr/testify/assert"
)

func TestHealthySrv(t *testing.T) {
	var config = `
    {
        "input": [
           {  
               "type": "gamehealthysrv",

               "redis_host": "localhost:6379",
               "redis_password": "",
               "redis_db": 0,

               "influx_url": "http://localhost:8086",
               "influx_user": "",
               "influx_passwd": "",

               "log_level": 5
           }
        ]
    }
    `
	srv, err := utils.LoadFromString(config)
	assert.NoError(t, err)
	srv.RunInputs()

	// stay alive
	time.Sleep(600 * time.Second)
	srv.StopInputs()

}
