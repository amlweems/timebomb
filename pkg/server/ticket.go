package server

import (
	"crypto/rand"
	"math/big"
)

const letterBytes = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randInt(n int) int {
	max := big.NewInt(int64(n))
	result, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return int(result.Int64())
}

func ticket() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[randInt(len(letterBytes))]
	}
	return string(b)
}
