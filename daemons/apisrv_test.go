package daemons

import (
	"testing"
	"time"
	"zonst/qipai/logagent/utils"

	"github.com/stretchr/testify/assert"
)

func TestHealthyAPISrv(t *testing.T) {
	var config = `
    {
        "input": [
           {  
               "type": "gamehealthyapisrv",

			   "host": "0.0.0.0:5000",

               "redis_host": "localhost:6379",
               "redis_password": "",
               "redis_db": 0,

			   "elastic_urls": ["http://10.0.10.46:9200"],
				
               "log_level": 5,
			   "accesskeys": [
					"vB9zXv6H0Pkzb",
					"OnamYpBVSRZHd",

					"ySZFbntnmLKBq",
					"yn9oyzWzgtBFf" 
               ]
           }
        ]
    }
    `
	srv, err := utils.LoadFromString(config)
	assert.NoError(t, err)
	srv.RunInputs()

	// stay alive
	time.Sleep(600 * time.Second)
}
