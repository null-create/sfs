package files

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func KbToMb(kb float64) float64 {
	return kb / 1024.0
}

// Generate a random integer in the range [1, n)
func RandInt(limit int) int {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	num := rand.Intn(limit)
	if num == 0 {
		return 1
	}
	return num
}

// creates a new UUID string
func NewUUID() string {
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("[ERROR] Could not generate UUID:\n%v\n", err)
	}
	return uuid.String()
}
