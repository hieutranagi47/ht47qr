package domain

import (
	"crypto/rand"
	"errors"
)

const Base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type CodeGenerator interface{ Generate() (string, error) }

type RandomCodeGenerator struct{}

func (RandomCodeGenerator) Generate() (string, error) {
	code := make([]byte, 8)
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		code[i] = Base62Alphabet[int(b)%len(Base62Alphabet)]
	}
	if len(code) != 8 {
		return "", errors.New("invalid generated code")
	}
	return string(code), nil
}
