package utils

import (
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
)

func GenerateUUID() string {
	id := uuid.New().String()

	return id
}

func ValidSolanaAddress(publicAddress string) error {
	_, err := solana.PublicKeyFromBase58(publicAddress)

	if err != nil {
		log.Printf("error occured while trying to parse the publicAddress, %v\n", err)
		return err
	}

	return nil
}