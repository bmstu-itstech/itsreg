package jwtauth

import "os"

var (
	secret = os.Getenv("JWT_SECRET")
)
