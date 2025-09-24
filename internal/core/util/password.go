package util

import "golang.org/x/crypto/bcrypt"

func GenerateEncrypt(password string) (string, error) {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(encrypted), nil
}

func ComparePassword(password, encrypted string) error {
	return bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(password))
}
