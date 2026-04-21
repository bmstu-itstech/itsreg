package nanoid

import gonanoid "github.com/matoous/go-nanoid/v2"

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewNanoID(l int) string {
	return gonanoid.MustGenerate(alphabet, l)
}
