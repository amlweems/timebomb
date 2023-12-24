package engine

import (
	"crypto/rand"
	"math/big"
)

func randN(n int) int {
	randNum, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(randNum.Int64())
}

func shuffle[T any](a []T) []T {
	b := a[:]
	for i := len(b) - 1; i > 0; i-- {
		j := randN(i + 1)
		b[i], b[j] = b[j], b[i]
	}
	return b
}
