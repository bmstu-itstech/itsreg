package uuid

import "github.com/jaevor/go-nanoid"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
const length = 8

var Generate = nanoid.MustCustomASCII(alphabet, length)
