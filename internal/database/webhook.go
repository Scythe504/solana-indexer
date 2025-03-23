package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) CreateWebhook(name string, txnType IndexingStrategy, address string) error {

	var (
		heliusApiKey        = os.Getenv("HELIUS_API_KEY")
		heliusApiUrl        = os.Getenv("HELIUS_API_URL")
		heliusWebhookSecret = os.Getenv("HELIUS_WEBHOOK_SECRET")
		publicUrl           = os.Getenv("PUBLIC_URL")
	)

	// TODO-Need to check whethe the address is a wallet address or token/NFT address
	publicKey, err := solana.PublicKeyFromBase58(address)

	if err != nil {
		log.Println("Error parsing public key: ", err)
		return err
	}

	body := map[string]interface{}{
		"webhookURL":  fmt.Sprintf("%s/webhook/%s", publicUrl, name),
		"webhookType": "enhanced",
		"transactionTypes": []IndexingStrategy{
			txnType,
		},
		"accountAddresses": []solana.PublicKey{publicKey},
		"txnStatus":        utils.TxnStatusSuccess,
	}

	// Only add auth header if secret is available
	if heliusWebhookSecret != "" {
		body["authHeader"] = fmt.Sprintf("Bearer %s", heliusWebhookSecret)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling webhook request body: %v", err)
		return err
	}

	url := fmt.Sprintf("%s/webhooks?api-key=%s", heliusApiUrl, heliusApiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending webhook creation request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Helius API error: %s - %s", resp.Status, string(respBody))
		return err
	}
	jsonResp := &utils.HeliusWebhookCreateResponse{}
	if err = json.Unmarshal(respBody, &jsonResp); err != nil {
		log.Println("Error occured while parsing json for create webhook: ", err)
		return err
	}

	var (
		webhookUuid = utils.GenerateUUID()
		now         = time.Now()
	)

	webhookConfig := HeliusWebhookConfig{
		Id:           webhookUuid,
		WebhookName:  name,
		WebhookId:    jsonResp.WebhookId,
		AddressCount: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = s.db.Exec(`
		INSERT INTO helius_webhook_config (
			id,
			webhook_name,
			webhook_id,
			address_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, webhookConfig.Id,
		webhookConfig.WebhookName,
		webhookConfig.WebhookId,
		webhookConfig.AddressCount,
		webhookConfig.CreatedAt,
		webhookConfig.UpdatedAt,
	)

	if err != nil {
		log.Println("Error occured while inserting into helius webhook config table: ", err)
		return err
	}

	return nil
}

func (s *service) GetAllWebhooks() ([]HeliusWebhookConfig, error) {
	var heliusWebhookCfg []HeliusWebhookConfig

	rows, err := s.db.Query(`SELECT * FROM HeliusWebhookConfig`)
	if err != nil {
		log.Println("Error occured while fetching webhookrecords", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cfg HeliusWebhookConfig

		err := rows.Scan(
			&cfg.Id,
			&cfg.WebhookName,
			&cfg.WebhookId,
			&cfg.AddressCount,
			&cfg.CreatedAt,
			&cfg.UpdatedAt,
		)

		if err != nil {
			log.Println("Error scanning webhook config: ", err)
			return nil, err
		}

		heliusWebhookCfg = append(heliusWebhookCfg, cfg)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating through webhook configs: ", err)
		return nil, err
	}

	return heliusWebhookCfg, nil
}

func (s *service) GetWebhookConfigByName(name string) (HeliusWebhookConfig, error) {
	var cfg HeliusWebhookConfig

	err := s.db.QueryRow(`
		SELECT (*)
		FROM helius_webhook_config 
		WHERE webhook_name = $1
	`, name).Scan(
		&cfg.Id,
		&cfg.WebhookName,
		&cfg.WebhookId,
		&cfg.AddressCount,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, fmt.Errorf("webhook config with name %s not found", name)
		}
		log.Println("Error querying webhook config by name:", err)
		return cfg, err
	}


	return cfg, nil
}