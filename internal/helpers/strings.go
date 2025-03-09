package helpers

import (
	"math/rand"
	"time"
)

func Shuffle(slice []string) {
	rand.NewSource(time.Now().UnixNano()) // Ensure randomness

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}
