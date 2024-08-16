package client

import (
	"math/rand"
)

// Generate a pseudo-random integer in the range [0, n)
func RandInt(limit int) int {
	return rand.Intn(limit)
}
