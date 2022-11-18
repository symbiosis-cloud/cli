package util

import (
	"math/rand"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyz01234567890"

func RandomString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
