package test

import (
	"fmt"
	"math/rand"
	"time"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var characters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randomString(prefix string, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characters[r.Intn(len(characters))]
	}

	return fmt.Sprintf("%s-%s", prefix, string(b))
}
