package steps

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

type cryptoReader struct {
	reader    io.ReadCloser
	psk       []byte
	chunkSize int
	currChunk []byte
}

func (r *cryptoReader) Read(b []byte) (int, error) {
	if len(r.currChunk) == 0 {
		p := make([]byte, r.chunkSize)

		n, err := io.ReadFull(r.reader, p)
		if err != nil && err != io.ErrUnexpectedEOF {
			return n, err
		}

		unencryptedChunk, err := decryptObjectContent(r.psk, p[:n])
		if err != nil {
			return 0, err
		}

		r.currChunk = unencryptedChunk
	}

	var n int
	if len(r.currChunk) >= len(b) {
		copy(b, r.currChunk[:len(b)])
		n = len(b)
		r.currChunk = r.currChunk[len(b):]
	} else {
		copy(b, r.currChunk)
		n = len(r.currChunk)
		r.currChunk = nil
	}

	return n, nil
}

func decryptObjectContent(psk []byte, encryptedBytes []byte) ([]byte, error) {
	block, err := aes.NewCipher(psk)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCFBDecrypter(block, psk)

	unencryptedBytes := make([]byte, len(encryptedBytes))
	stream.XORKeyStream(unencryptedBytes, encryptedBytes)

	return unencryptedBytes, nil
}
