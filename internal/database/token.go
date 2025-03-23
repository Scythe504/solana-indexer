package database

import (
	"log"

	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) GetSubscriptionsByWebhookId(webhookId string) ([]TokenSubscriptionLookup, error) {
	var token_subscriptions []TokenSubscriptionLookup

	rows, err := s.db.Query(`
		SELECT (
			id, 
			token_address,
			user_id, 
			table_name,
			helius_webhook_id, 
			last_updated
		) FROM token_subscription_lookup
		WHERE helius_webhook_id = $1
	`, webhookId)
	if err != nil {
		log.Println("Error occured while fetching TokenSubscriptionLookup table")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var token_subscription TokenSubscriptionLookup

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

func (s *service) RegisterToken(publicAddress string) error {
	if err := utils.ValidSolanaAddress(publicAddress); err != nil  {
		log.Println("Invalid solana address")
		return err
	}

	

	return nil
}