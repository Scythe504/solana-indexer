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

type AddressType string

const (
	WALLET AddressType = "wallet"
	TOKEN  AddressType = "token"
	NFT    AddressType = "nft"
)

func ValidSolanaAddress(publicAddress string) error {
	_, err := solana.PublicKeyFromBase58(publicAddress)

	if err != nil {
		log.Printf("error occured while trying to parse the publicAddress, %v\n", err)
		return err
	}

	return nil
}

// func GetSolanaAddressType(publicAddress string) {
// 	client := rpc.New(rpc.MainNetBeta_RPC)
	
	
// }
// func GetSolanaAddressData(publicAddress string) (*rpc.DataBytesOrJSON, error) {
// 	client := rpc.New(rpc.MainNetBeta_RPC)

// 	address, _ := solana.PublicKeyFromBase58(publicAddress)

// 	account, err := client.GetAccountInfo(
// 		context.Background(),
// 		address,
// 	)

// 	if err != nil {
// 		log.Printf("Failed to get account info: %v\n", err)
// 		return nil, err
// 	}

// 	return account.Value.Data, nil
// }
