package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/scythe504/solana-indexer/internal/database"
	"github.com/scythe504/solana-indexer/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

func ParseBodyAndPushToProducer(m *KafkaClientManager, body []byte, receiverName string) error {

	var jsonResp []WebhookPayload

	err := json.Unmarshal(body, &jsonResp)

	if err != nil {
		log.Println("Error occured while parsing body, Invalid Json, err: ", err)
		return err
	}

	kafkaClient, err := m.GetClient()
	if err != nil {
		log.Printf("Failed to create Kafka producer: %v", err)
		return err
	}
	defer kafkaClient.Close()

	if err = m.ProduceWebhookPayload(kafkaClient, jsonResp, receiverName); err != nil {
		log.Printf("Error occured while trying to produce, err: %v", err)
		return err
	}

	return nil
}

func StoreRecordForInterestedUsers(record *kgo.Record) error {
	recordValue := record.Value
	receiverName := string(record.Key)

	var jsonResp []WebhookPayload

	if err := json.Unmarshal(recordValue, &jsonResp); err != nil {
		log.Println("Error occured while parsing record value, err: ", err)
		return err
	}

	var prevTxnType database.IndexingStrategy

	for _, resp := range jsonResp {

		var finalInterestedSubscriptions []database.SubscriptionLookup
		addressLookupSet := make(AddressSet)
		var txnType database.IndexingStrategy
		for _, strats := range database.StrategyTransactionTypes {
			if slices.Contains(strats, resp.Type) {
				txnType = database.IndexingStrategy(resp.Type)
				prevTxnType = txnType
				break
			}
		}

		if txnType == "" {
			continue
		}

		var subscriptions []database.SubscriptionLookup
		var err error
		if prevTxnType != txnType {
			subscriptions, err = database.Service.GetSubscriptionsByTxnType(database.New(), txnType, receiverName)
			if err != nil {
				log.Printf("Error occurred while fetching subscriptions for txnType: %s, receiverName: %s, Error: %v", txnType, receiverName, err)
				return err
			}
			prevTxnType = txnType
		}

		for _, accountData := range resp.AccountData {
			if !addressLookupSet.Contains(accountData.Account) {
				addressLookupSet[accountData.Account] = true
			}
			for _, tokenBalanceChanges := range accountData.TokenBalanceChanges {
				if !addressLookupSet.Contains(tokenBalanceChanges.Mint) {
					addressLookupSet[tokenBalanceChanges.Mint] = true
				}
				if !addressLookupSet.Contains(tokenBalanceChanges.UserAccount) {
					addressLookupSet[tokenBalanceChanges.UserAccount] = true
				}

				if !addressLookupSet.Contains(tokenBalanceChanges.TokenAccount) {
					addressLookupSet[tokenBalanceChanges.TokenAccount] = true
				}
			}
		}

		if !addressLookupSet.Contains(resp.FeePayer) {
			addressLookupSet[resp.FeePayer] = true
		}

		for _, val := range resp.Instructions {
			for _, acc := range val.Accounts {
				if !addressLookupSet.Contains(acc) {
					addressLookupSet[acc] = true
				}

			}
			for _, inner := range val.InnerInstructions {
				for _, account := range inner.Accounts {
					if !addressLookupSet.Contains(account) {
						addressLookupSet[account] = true
					}
				}
			}
		}

		for _, val := range resp.NativeTransfers {
			if !addressLookupSet.Contains(val.FromUserAccount) {
				addressLookupSet[val.FromUserAccount] = true
			}
			if !addressLookupSet.Contains(val.ToUserAccount) {
				addressLookupSet[val.ToUserAccount] = true
			}
		}

		for _, tokenTransfer := range resp.TokenTransfers {
			if !addressLookupSet.Contains(tokenTransfer.FromUserAccount) {
				addressLookupSet[tokenTransfer.FromUserAccount] = true
			}
			if !addressLookupSet.Contains(tokenTransfer.ToUserAccount) {
				addressLookupSet[tokenTransfer.ToUserAccount] = true
			}
			if !addressLookupSet.Contains(tokenTransfer.FromTokenAccount) {
				addressLookupSet[tokenTransfer.FromTokenAccount] = true
			}
			if !addressLookupSet.Contains(tokenTransfer.ToTokenAccount) {
				addressLookupSet[tokenTransfer.ToTokenAccount] = true
			}
			if !addressLookupSet.Contains(tokenTransfer.Mint) {
				addressLookupSet[tokenTransfer.Mint] = true
			}
		}

		for _, subscription := range subscriptions {
			for key := range addressLookupSet {
				if subscription.TokenAddress == key {
					finalInterestedSubscriptions = append(finalInterestedSubscriptions, subscription)
					break
				}
			}
		}

		IndexDataForUsers(finalInterestedSubscriptions, resp)
	}

	return nil
}

func IndexDataForUsers(subscriptions []database.SubscriptionLookup, jsonPayload WebhookPayload) {
	var prevUserId string
	var dbConfig *database.UserDatabaseCredential
	var err error
	for _, subscription := range subscriptions {
		if prevUserId != subscription.UserId {
			dbConfig, err = database.Service.GetDatabaseConfig(database.New(), subscription.UserId)
			if err != nil {
				log.Printf("Error occured while fetching for database config for user: %s, err: %v", subscription.UserId, err)
				continue
			}
			prevUserId = subscription.UserId
		}

		var (
			dbUrl   = dbConfig.ConnectionString
			connStr string
		)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if dbUrl != nil {
			connStr = *dbUrl
		} else {
			connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", *dbConfig.User, *dbConfig.Password, *dbConfig.Host, *dbConfig.Port, *dbConfig.DatabaseName, *dbConfig.SSLMode)
		}
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			log.Printf("Error occured while connecting to database, userId: %s\n", dbConfig.UserId)
			continue
		}
		defer db.Close()
		if err = db.PingContext(ctx); err != nil {
			log.Printf("Error occured while connecting to database, userId: %s\n", dbConfig.UserId)
			continue
		}
		InsertPayloadInUserDatabase(ctx, db, jsonPayload, dbConfig.UserId, subscription.TableName)

	}
}

func InsertPayloadInUserDatabase(ctx context.Context, db *sql.DB, payload WebhookPayload, userId string, tableName string) {
	jsonBlob, err := json.Marshal(payload)

	if err != nil {
		log.Println("Failed to encode into json for userId: ", userId, err)
		return
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

	db_uuid := utils.GenerateUUID()
	if err != nil {
		log.Printf("Failed to start transaction")
	}
	defer tx.Rollback()
	// Store raw json obj
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id VARCHAR(255) NOT NULL PRIMARY KEY, jsonData JSONB)", tableName)
	_, err = tx.ExecContext(ctx, query)

	if err != nil {
		log.Println("Failed to create tables cannot proceed, userId: ", userId, err)
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (id, jsonData) VALUES ($1, $2)", tableName)

	_, err = tx.ExecContext(ctx, insertQuery, db_uuid, string(jsonBlob))

	if err != nil {
		log.Println("Failed to execute database query: ", userId, err)
	}

	if err = CreateAndInsertNormalizedData(tx, payload, db_uuid); err != nil {
		log.Println("Failed to insert normalized data, err: ", err)
	}

	if err = tx.Commit(); err != nil {
		log.Println("Failed to commit transaction, changes will be rolled back")
	}
}

func CreateAndInsertNormalizedData(tx *sql.Tx, payload WebhookPayload, payloadID string) error {
	tableCreationQueries := []string{
		`CREATE TABLE IF NOT EXISTS webhook_payloads (
			id VARCHAR(255) PRIMARY KEY,
			signature VARCHAR(255) NOT NULL,
			slot BIGINT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			source VARCHAR(100),
			fee INTEGER,
			fee_payer VARCHAR(255),
			transaction_type VARCHAR(50),
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS account_data (
			id SERIAL PRIMARY KEY,
			payload_id VARCHAR(255) REFERENCES normalized_webhook_payloads(id),
			account VARCHAR(255) NOT NULL,
			native_balance_change NUMERIC(20,0)
		)`,
		`CREATE TABLE IF NOT EXISTS token_balance_changes (
			id SERIAL PRIMARY KEY,
			account_data_id INTEGER REFERENCES account_data(id),
			mint VARCHAR(255) NOT NULL,
			token_account VARCHAR(255) NOT NULL,
			user_account VARCHAR(255) NOT NULL,
			token_amount VARCHAR(100) NOT NULL,
			decimals SMALLINT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS native_transfers (
			id SERIAL PRIMARY KEY,
			payload_id VARCHAR(255) REFERENCES webhook_payloads(id),
			from_user_account VARCHAR(255) NOT NULL,
			to_user_account VARCHAR(255) NOT NULL,
			amount NUMERIC(20,0) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS token_transfers (
			id SERIAL PRIMARY KEY,
			payload_id VARCHAR(255) REFERENCES normalized_webhook_payloads(id),
			from_token_account VARCHAR(255) NOT NULL,
			from_user_account VARCHAR(255) NOT NULL,
			to_token_account VARCHAR(255) NOT NULL,
			to_user_account VARCHAR(255) NOT NULL,
			mint VARCHAR(255) NOT NULL,
			token_amount NUMERIC(20,8),
			token_standard VARCHAR(50)
		)`,
	}

	for _, query := range tableCreationQueries {
		_, err := tx.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create normalized tables: %v", err)
		}
	}

	// Insert payload metadata
	payloadQuery := `
		INSERT INTO normalized_webhook_payloads 
		(id, signature, slot, timestamp, source, fee, fee_payer, transaction_type, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := tx.Exec(
		payloadQuery,
		payloadID,
		payload.Signature,
		payload.Slot,
		time.Unix(payload.Timestamp, 0),
		payload.Source,
		payload.Fee,
		payload.FeePayer,
		payload.Type,
		payload.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to insert webhook payload: %v", err)
	}

	// Insert account data
	for _, accountData := range payload.AccountData {
		accountQuery := `
			INSERT INTO normalized_account_data 
			(payload_id, account, native_balance_change)
			VALUES ($1, $2, $3)
			RETURNING id
		`
		var accountDataID int
		err = tx.QueryRow(
			accountQuery,
			payloadID,
			accountData.Account,
			accountData.NativeBalanceChange.String(),
		).Scan(&accountDataID)
		if err != nil {
			return fmt.Errorf("failed to insert account data: %v", err)
		}

		// Insert token balance changes
		for _, tokenChange := range accountData.TokenBalanceChanges {
			tokenChangeQuery := `
				INSERT INTO normalized_token_balance_changes
				(account_data_id, mint, token_account, user_account, token_amount, decimals)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = tx.Exec(
				tokenChangeQuery,
				accountDataID,
				tokenChange.Mint,
				tokenChange.TokenAccount,
				tokenChange.UserAccount,
				tokenChange.RawTokenAmount.TokenAmount,
				tokenChange.RawTokenAmount.Decimals,
			)
			if err != nil {
				return fmt.Errorf("failed to insert token balance changes: %v", err)
			}
		}
	}

	// Insert native transfers
	for _, transfer := range payload.NativeTransfers {
		nativeTransferQuery := `
			INSERT INTO normalized_native_transfers
			(payload_id, from_user_account, to_user_account, amount)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(
			nativeTransferQuery,
			payloadID,
			transfer.FromUserAccount,
			transfer.ToUserAccount,
			transfer.Amount.String(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert native transfers: %v", err)
		}
	}

	// Insert token transfers
	for _, transfer := range payload.TokenTransfers {
		tokenTransferQuery := `
			INSERT INTO normalized_token_transfers
			(payload_id, from_token_account, from_user_account, 
			 to_token_account, to_user_account, mint, 
			 token_amount, token_standard)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.Exec(
			tokenTransferQuery,
			payloadID,
			transfer.FromTokenAccount,
			transfer.FromUserAccount,
			transfer.ToTokenAccount,
			transfer.ToUserAccount,
			transfer.Mint,
			transfer.TokenAmount,
			transfer.TokenStandard,
		)
		if err != nil {
			return fmt.Errorf("failed to insert token transfers: %v", err)
		}
	}

	return nil
}
