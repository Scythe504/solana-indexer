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
	"slices"
	"time"

	"github.com/gagliardetto/solana-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/scythe504/solana-indexer/internal/utils"
)

var (
	heliusApiKey        = os.Getenv("HELIUS_API_KEY")
	heliusApiUrl        = os.Getenv("HELIUS_API_URL")
	heliusWebhookSecret = os.Getenv("HELIUS_WEBHOOK_SECRET")
	publicUrl           = os.Getenv("PUBLIC_URL")
)

func (s *service) CreateWebhook(name string, txnType []IndexingStrategy, address string) error {

	// TODO-Need to check whether the address is a wallet address or token/NFT address
	publicKey, err := solana.PublicKeyFromBase58(address)

	if err != nil {
		log.Println("Error parsing public key: ", err)
		return err
	}

	body := map[string]interface{}{
		"webhookURL":       fmt.Sprintf("%s/webhook/%s", publicUrl, name),
		"webhookType":      "enhanced",
		"transactionTypes": txnType,
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
	jsonResp := &utils.HeliusWebhookResponse{}
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
		SELECT *
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

func (s *service) CreateOrUpdateWebhook(address string, txnType []IndexingStrategy) error {
	heliusConfig, err := s.GetAllWebhooks()

	if err != nil {
		log.Println("Error occured while fetching webhooks", err)
		return err
	}

	var (
		name = fmt.Sprintf("webhook-%d", len(heliusConfig))
	)
	switch length := len(heliusConfig); length {
	case 0:
		if err := s.CreateWebhook(name, txnType, address); err != nil {
			log.Println("Error occured while creating webhooks", err)
			return err
		}
	default:
		if err := s.UpdateWebhook(heliusConfig, address, txnType); err != nil {
			log.Println("Error occured while updating webhook")
			return err
		}
	}

	return nil
}

func (s *service) UpdateWebhook(heliusConfig []HeliusWebhookConfig, address string, txnType []IndexingStrategy) error {
	var (
		authHeader = fmt.Sprintf("Bearer %s", os.Getenv("HELIUS_WEBHOOK_SECRET"))
		txnStatus  = utils.TxnStatusSuccess
	)

	for _, config := range heliusConfig {
		if config.AddressCount >= 99999 {
			continue
		}

		webhookConfig, err := s.GetCurrentWebhookConfig(config.WebhookId)

		if err != nil {
			log.Println("Failed to get current webhook config from helius: ", err)
			return err
		}

		body := map[string]interface{}{
			"webhookURL":  fmt.Sprintf("%s/webhook/%s", publicUrl, config.WebhookName),
			"webhookType": "enhanced",
			"txnStatus":   txnStatus,
			"authHeader":  authHeader,
		}

		uniqueTxnTypes := make(map[string]bool)
		for _, txnType := range webhookConfig.TransactionTypes {
			uniqueTxnTypes[txnType] = true
		}

		var notPresentTxnTypes []string
		for _, newType := range txnType {
			txnTypeStr := string(newType)
			if _, exists := uniqueTxnTypes[txnTypeStr]; !exists {
				notPresentTxnTypes = append(notPresentTxnTypes, txnTypeStr)
				uniqueTxnTypes[txnTypeStr] = true
			}
		}

		addressExists := slices.Contains(webhookConfig.AccountAddresses, address)

		if len(notPresentTxnTypes) == 0 && addressExists {
			return nil // No changes needed
		}

		if !addressExists {
			body["accountAddresses"] = append(webhookConfig.AccountAddresses, address)
		}

		if len(notPresentTxnTypes) > 0 {
			body["transactionTypes"] = append(webhookConfig.TransactionTypes, notPresentTxnTypes...)
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Printf("Error marshaling webhook request body: %v", err)
			return err
		}

		url := fmt.Sprintf("%s/webhooks?api-key=%s", heliusApiUrl, heliusApiKey)
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
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

		jsonResp := &utils.HeliusWebhookResponse{}
		if err = json.Unmarshal(respBody, &jsonResp); err != nil {
			log.Println("Error occured while parsing json for create webhook: ", err)
			return err
		}

		_, err = s.db.Exec(`
			UPDATE helius_webhook_config
			SET address_count = $1
			WHERE helius_webhook_id = $2
		`, config.AddressCount+1, jsonResp.WebhookId)
		if err != nil {
			log.Println("Failed to update the address count")
		}
		break
	}

	return nil
}

func (s *service) GetCurrentWebhookConfig(webhookId string) (*utils.HeliusWebhookResponse, error) {
	url := fmt.Sprintf("%s/webhooks/%s?api-key=%s", heliusApiUrl, webhookId, heliusApiKey)

	resp, err := http.Get(url)

	if err != nil {
		log.Println("Failed to fetch url")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Request Failed for fetching Webhook Config")
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("Failed to read response body")
		return nil, err
	}
	defer resp.Body.Close()

	jsonResp := &utils.HeliusWebhookResponse{}
	if err = json.Unmarshal(respBody, &jsonResp); err != nil {
		log.Println("Failed to Unmarshal into WebhookResponse struct: ", err)
		return nil, err
	}

	return jsonResp, nil
}
