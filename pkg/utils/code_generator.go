package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRandomCode generates a random n-digit code.
func GenerateRandomCode(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%0*d", n, r.Intn(pow(10, n)))
}

func pow(a, b int) int {
	p := 1
	for b > 0 {
		p *= a
		b--
	}
	return p
}
