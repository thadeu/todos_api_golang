package factory

import (
	fab "github.com/Goldziher/fabricator"
	"golang.org/x/crypto/bcrypt"
)

func NewUser[T any](customData ...map[string]any) T {
	instance := fab.New(*new(T))

	if len(customData) > 0 {
		hasEncryptedPassword := false

		for _, data := range customData {
			if _, exists := data["EncryptedPassword"]; exists {
				hasEncryptedPassword = true
				break
			}
		}

		if !hasEncryptedPassword {
			encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)

			customData = append(customData, map[string]any{
				"EncryptedPassword": string(encryptedPassword),
			})
		}
	}

	return instance.Build(customData...)
}
