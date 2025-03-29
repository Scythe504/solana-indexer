package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) GetSubscriptionsByWebhookId(webhookId string) ([]SubscriptionLookup, error) {
	var token_subscriptions []SubscriptionLookup

	rows, err := s.db.Query(`
		SELECT 
			id,
			token_address,
			user_id, 
			table_name,
			helius_webhook_id, 
			last_updated
		FROM subscription_lookup
		WHERE helius_webhook_id = $1
	`, webhookId)
	if err != nil {
		log.Println("Error occured while fetching TokenSubscriptionLookup table")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var token_subscription SubscriptionLookup

		err := rows.Scan(
			&token_subscription.Id,
			&token_subscription.TokenAddress,
			&token_subscription.UserId,
			&token_subscription.Strategy,
			&token_subscription.TableName,
			&token_subscription.HeliusWebhookId,
			&token_subscription.LastUpdated,
		)

		if err != nil {
			log.Println("Error while scanning token_subscription_lookup rows: ", err)
			return nil, err
		}

		token_subscriptions = append(token_subscriptions, token_subscription)
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating through webhook configs: ", err)
		return nil, err
	}

	return token_subscriptions, nil
}

func (s *service) RegisterAddress(tx *sql.Tx, token AddressRegistery) error {
	if err := utils.ValidSolanaAddress(token.TokenAddress); err != nil {
		log.Println("Invalid solana address")
		return err
	}

	if tx != nil {
		now := time.Now()
		_, err := tx.Exec(`
		INSERT INTO address_registry (
			id,
			token_address,
			token_name,
			token_symbol,
			created_at,
			last_fetched_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, token.Id, token.TokenAddress, token.TokenName, token.TokenSymbol, now, now)

		if err != nil {
			log.Printf("Error occured while trying to register token %s, Error: %v\n", token.TokenAddress, err)
			return err
		}
		return nil
	} else {

		now := time.Now()
		_, err := s.db.Exec(`
		INSERT INTO address_registery (
			id,
			token_address,
			token_name,
			token_symbol,
			created_at,
			last_fetched_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, token.Id, token.TokenAddress, token.TokenName, token.TokenSymbol, now, now)

		if err != nil {
			log.Printf("Error occured while trying to register token %s, Error: %v\n", token.TokenAddress, err)
			return err
		}

		return nil
	}
}

// func (s *service) GetAddressData(publicAddress string) (*AddressRegistery, error) {
// 	var token AddressRegistery

// 	if err := utils.ValidSolanaAddress(publicAddress); err != nil {
// 		log.Printf("Invalid Solana Address: %v\n", err)
// 		return nil, err
// 	}

// 	tokenData, err := utils.GetSolanaAddressData(publicAddress)

// }

// func (s *service) SubscribeToAddress(userId string) error {

// 	return nil
// }

func (s *service) GetSubscriptionsByAddressAndTxnType(address string, txnType IndexingStrategy, receiverName string) ([]SubscriptionLookup, error) {
	var subscriptions []SubscriptionLookup

	var heliusConfig HeliusWebhookConfig
	err := s.db.QueryRow(`SELECT 
		id,
		webhook_id
	  FROM helius_webhook_config
	  WHERE webhook_name = $1
	`, receiverName).Scan(
		&heliusConfig.Id,
		&heliusConfig.WebhookId,
	)

	if err != nil {
		log.Printf("Error occured while trying to get helius configs: %v", err)
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT id,
			token_address,
			strategy,
			table_name,
			helius_webhook_id,
			last_updated
		 FROM subscription_lookup
		  WHERE (
			token_address = $1,
			strategy = $2
		  )
	`, address, txnType)

	if err != nil {
		log.Println("Query failed for subscription_lookup", err)
		return nil, err
	}

	for rows.Next() {
		var subscription SubscriptionLookup

		err := rows.Scan(
			&subscription.Id,
			&subscription.TokenAddress,
			&subscription.UserId,
			&subscription.Strategy,
			&subscription.TableName,
			&subscription.HeliusWebhookId,
			&subscription.LastUpdated,
		)

		if err != nil {
			log.Println("Some error occured while getting the subscriptions")
			return nil, err
		}

		subscriptions = append(subscriptions, subscription)
	}
	defer rows.Close()

	return subscriptions, nil
}

func (s *service) CheckIfSubscriptionsAlreadyExistByUser(userId string, tokenAddress string) (bool, error) {
	row, err := s.db.Query(`
		SELECT 
			tokenAddress
		FROM subscriptions
		WHERE user_id = $1 AND token_address = $2
	`, userId, tokenAddress)

	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("error occured while fetching the userId and tokenAddress from subscriptions")
	}

	if row != nil {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *service) GetAddressFromRegistery(address string) (*AddressRegistery, error) {
	var reg AddressRegistery
	err := s.db.QueryRow(`
		SELECT 
			id, 
			token_address,
			token_name,
			token_symbol,
			created_at,
			last_fetched_at
		 FROM address_registry 
		  WHERE token_address = $1
	`, address).Scan(
		&reg.Id,
		&reg.TokenAddress,
		&reg.TokenName,
		&reg.TokenSymbol,
		&reg.CreatedAt,
		&reg.LastFetchedAt,
	)

	if err != nil {
		log.Println("Failed to get the token from registery")
		return nil, err
	}

	return &reg, nil
}

func (s *service) GetSubscriptionsByTxnType(txnType IndexingStrategy, receiverName string) ([]SubscriptionLookup, error) {
	var subscriptions []SubscriptionLookup

	var heliusConfig HeliusWebhookConfig
	err := s.db.QueryRow(`SELECT 
		id,
		webhook_id
	 FROM helius_webhook_config
	  WHERE webhook_name = $1
	`, receiverName).Scan(
		&heliusConfig.Id,
		&heliusConfig.WebhookId,
	)

	if err != nil {
		log.Printf("Error occured while trying to get helius configs: %v", err)
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT 
			id,
			token_address,
			strategy,
			table_name,
			helius_webhook_id,
			last_updated
		 FROM subscription_lookup
		  WHERE ( 
			strategy = $1,
			helius_webhook_id = $2
		   )
	`, txnType, heliusConfig.WebhookId)

	if err != nil {
		log.Println("Query failed for subscription_lookup", err)
		return nil, err
	}

	for rows.Next() {
		var subscription SubscriptionLookup

		err := rows.Scan(
			&subscription.Id,
			&subscription.TokenAddress,
			&subscription.UserId,
			&subscription.Strategy,
			&subscription.TableName,
			&subscription.HeliusWebhookId,
			&subscription.LastUpdated,
		)

		if err != nil {
			log.Println("Some error occured while getting the subscriptions")
			return nil, err
		}

		subscriptions = append(subscriptions, subscription)
	}
	defer rows.Close()

	return subscriptions, nil
}
func (s *service) CreateSubscription(tokenAddress string, strats []IndexingStrategy, userId string) error {
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Println("Failed to begin a transaction: ", err)
		return err
	}

	addressIsPresent, err := s.CheckIfSubscriptionsAlreadyExistByUser(userId, tokenAddress)

	if err != nil {
		log.Println("Failed to check if subscription exists")
		return err
	}

	if addressIsPresent {
		return fmt.Errorf("indexing for this address already exists, please edit the txn types if you need to index more data for the given address")
	}

	defer tx.Rollback()
	token, err := s.GetAddressFromRegistery(tokenAddress)

	var addressReg *AddressRegistery

	// If token not found in registry
	if err == sql.ErrNoRows {
		log.Println("Token not found in registry, attempting to fetch")

		// Check if it's a token program account
		addressOwner, err := utils.GetSolanaAddressOwner(tokenAddress)
		if err != nil {
			log.Println("Failed to get owner of the account")
			return err
		}

		if solana.MustPublicKeyFromBase58(addressOwner) == solana.TokenProgramID {
			addressReg, err = FetchTokenDataFromHelius(tokenAddress)
			if err != nil {
				log.Println("Failed to fetch token data from helius")
				return err
			}

			err = s.RegisterAddress(tx, *addressReg)
			if err != nil {
				log.Println("Error occurred while registering token", err)
				return err
			}
		} else {
			return fmt.Errorf("invalid token address: not a token program account")
		}
	} else if err != nil {
		// Handle other potential errors
		log.Println("Unexpected error fetching from registry", err)
		return err
	}

	// Determine which token address to use
	finalTokenAddress := tokenAddress
	if token != nil {
		finalTokenAddress = token.TokenAddress
	} else if addressReg != nil {
		finalTokenAddress = addressReg.TokenAddress
	}

	var tableName string
	// Determine table name
	if addressReg != nil {
		tableName = addressReg.TokenName
	}
	if token != nil && token.TokenName != "" {
		tableName = token.TokenName
	}

	uuid := utils.GenerateUUID()
	now := time.Now()

	_, err = tx.Exec(`
		INSERT INTO subscriptions (
			id,
			user_id,
			token_address,
			indexing_strategy,
			table_name,
			created_at,
			updated_at,
			status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		uuid,
		userId,
		finalTokenAddress,
		strats,
		tableName,
		now,
		now,
		true,
	)
	if err != nil {
		log.Println("Failed to insert subscription in database:", err)
		return err
	}

	return tx.Commit()
}

func FetchTokenDataFromHelius(tokenAddress string) (*AddressRegistery, error) {
	// Helius API endpoint for token metadata
	url := fmt.Sprintf("%s/?api-key=%s", os.Getenv("HELIUS_RPC_URL"), os.Getenv("HELIUS_API_KEY"))

	// Prepare request payload
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      "test",
		"method":  "getAsset",
		"params": map[string]any{
			"id": tokenAddress,
		},
	}

	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("helius API returned non-200 status: %d, body: %s",
			resp.StatusCode, string(payload))
	}

	var tokenMetadata utils.HeliusRpcResponse
	if err := json.Unmarshal(payload, &tokenMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Println(tokenMetadata)

	// Check if we got any metadata
	if tokenMetadata.Result.Content.Metadata == nil {
		return nil, fmt.Errorf("no metadata found for token address %s", tokenAddress)
	}

	return &AddressRegistery{
		Id:            utils.GenerateUUID(),
		TokenAddress:  tokenAddress,
		TokenName:     tokenMetadata.Result.Content.Metadata.Name,
		TokenSymbol:   tokenMetadata.Result.Content.Metadata.Symbol,
		CreatedAt:     time.Now(),
		LastFetchedAt: &time.Time{},
	}, nil
}
