package encryption

import "crypto/rand"

type GenerateKey func() ([]byte, error)

func CreateKey() ([]byte, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
