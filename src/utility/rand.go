package utility

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	// Currently only these ports are whitelisted on AWS security group!
	StartPort = 30001
	EndPort   = 35000
	RangePort = EndPort - StartPort
)

// Returns a random port number in range [PortStart, PortEnd)
func RandPort() int {
	return StartPort + int(r.Intn(RangePort))
}
