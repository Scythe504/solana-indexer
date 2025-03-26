package database

import (
	"log"
	"time"

	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) GetSubscriptionsByWebhookId(webhookId string) ([]SubscriptionLookup, error) {
	var token_subscriptions []SubscriptionLookup

	rows, err := s.db.Query(`
		SELECT (
			id,
			token_address,
			user_id, 
			table_name,
			helius_webhook_id, 
			last_updated
		) FROM subscription_lookup
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

func (s *service) RegisterAddress(token AddressRegistery) error {
	if err := utils.ValidSolanaAddress(token.TokenAddress); err != nil {
		log.Println("Invalid solana address")
		return err
	}

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
	err := s.db.QueryRow(`SELECT (
		id,
		webhook_id
	) FROM helius_webhook_config
	  WHERE webhook_name = $1
	`, receiverName).Scan(
		&heliusConfig.Id,
		&heliusConfig.WebhookId,
	)

	if err != nil {
		log.Printf("Error occured while trying to get helius configs", err)
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT (
			id,
			token_address,
			strategy,
			table_name,
			helius_webhook_id,
			last_updated
		) FROM subscription_lookup
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

func (s *service) GetSubscriptionsByTxnType(txnType IndexingStrategy, receiverName string) ([]SubscriptionLookup, error) {
	var subscriptions []SubscriptionLookup

	var heliusConfig HeliusWebhookConfig
	err := s.db.QueryRow(`SELECT (
		id,
		webhook_id
	) FROM helius_webhook_config
	  WHERE webhook_name = $1
	`, receiverName).Scan(
		&heliusConfig.Id,
		&heliusConfig.WebhookId,
	)

	if err != nil {
		log.Printf("Error occured while trying to get helius configs", err)
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT (
			id,
			token_address,
			strategy,
			table_name,
			helius_webhook_id,
			last_updated
		) FROM subscription_lookup
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
