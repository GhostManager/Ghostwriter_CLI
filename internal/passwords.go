package internal

// Functions for generating random passwords

import (
	"crypto/rand"
	"log"
	"math/big"
	"strings"
)

// Generate a random password of the given length
// The password will be comprised of a-zA-Z0-9 and !@#$%^&*()_-+=/?<>.,
// Special characters exclude the following: '";:`~\/|
// Exclusions are to help avoid issues with escaping and breaking quotes in env files
func GenerateRandomPassword(pw_length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_-+=/?<>.,")
	var b strings.Builder
	for i := 0; i < pw_length; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			log.Fatalf("Failed to generate random number for password generation\n")
		}
		b.WriteRune(chars[nBig.Int64()])
	}
	return b.String()
}
