package modules

import "math/rand"

func GenerateRandomString(n int) string {
	symbols := "abcdefghijkl1234567890mnopqrstuvwxyz1234567890ABCDEFGHIJKL1234567890MNOPQRSTUVWXYZ1234567890"

	b := make([]byte, n)
	for i := range b {
		b[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(b)
}
