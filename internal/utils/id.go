package utils

import "github.com/google/uuid"

func GenerateUUID() string {
	id := uuid.New().String()

	return id
}
