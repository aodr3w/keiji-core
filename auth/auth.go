package auth

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/aodr3w/keiji-core/logging"
	"golang.org/x/crypto/bcrypt"
)

var log = logging.NewStdoutLogger()

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Password Encryption Error %v", err.Error())
		return "", err
	}
	return string(hashedPassword), nil
}

func IsValidPassword(hashedPassword []byte, password string) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
