/*
	DEFAULT CONFIG PARAM
*/
package master

import (
	"time"
)

const (
	VERSION = "0.0.1"

	DEFAULT_PORT = 10000

	DEFAULT_HOST = "0.0.0.0"

	DEFAULT_MAX_CONNECTIONS = (64 * 1024)

	DEFAULT_KEEPAlIVE = 300

	DEFAULT_CONNECT_TIMEOUT   = 2

	DEFAULT_ACKT_IMEOUT       = 20

	DEFAULT_TIMEOUT_RETRIES   = 3

	ACCEPT_MIN_SLEEP = 10 * time.Millisecond

	ACCEPT_MAX_SLEEP = 1 * time.Second

)
