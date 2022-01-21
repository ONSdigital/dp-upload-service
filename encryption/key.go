package encryption

import "crypto/rand"

type GenerateKey func() []byte

func CreateKey() []byte {
	key := make([]byte, 16)
	rand.Read(key) //nolint

	return key
}
