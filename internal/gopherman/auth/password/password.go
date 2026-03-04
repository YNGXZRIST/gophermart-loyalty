package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = bcrypt.DefaultCost

func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}
func Compare(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
