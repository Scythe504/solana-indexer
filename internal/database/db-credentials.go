package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) CreateDatabaseForUser(userId string, dbCred UserDatabaseCredential) error {
	var connString string

	if dbCred.ConnectionString == nil {
		connString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require&search_path=public", *dbCred.User, *dbCred.Password, *dbCred.Host, *dbCred.Port, *dbCred.DatabaseName)
	} else {
		connString = *dbCred.ConnectionString
	}

	db, err := sql.Open("pgx", connString)

	if err != nil {
		log.Printf("Error occured while connecting to userId:%s, \nDatabase: %s\n Error", userId, connString)
		return err
	}

	defer db.Close()

	err = db.Ping()

	if err != nil {
		log.Printf("Failed to ping database for userId:%s, Error: %v", userId, err)
		return err
	}

	now := time.Now()

	_, err = s.db.Exec(`
		INSERT INTO user_database_credential 
		(
			id,
			user_id,
			db_name,
			host,
			user,
			port,
			password, 
			ssl_mode, 
			connection_string,
			connection_limit,
			status,
			last_connected_at,
			created_at,
			updated_at, 
			error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, utils.GenerateUUID(),
		userId,
		dbCred.Host,
		dbCred.User,
		dbCred.Port,
		dbCred.Password,
		dbCred.SSLMode,
		dbCred.ConnectionString,
		dbCred.ConnectionLimit,
		dbCred.Status,
		now,
		now,
		now,
		nil,
	)

	if err != nil {
		log.Printf("error while inserting into user db credentials for userId: %s\n err: %v", userId, err)
		return err
	}
	return nil
}

func (s *service) GetDatabaseConfig(userId string) (*UserDatabaseCredential, error) {
	var databaseConfig UserDatabaseCredential

	err := s.db.QueryRow(`
		SELECT (
			id,
			user_id,
			db_name,
			host,
			port,
			user,
			password,
			ssl_mode,
			connection_string
		) FROM user_database_credential
		  WHERE user_id = $1
	`, userId).Scan(
		&databaseConfig.ID,
		&databaseConfig.UserId,
		&databaseConfig.DatabaseName,
		&databaseConfig.Host,
		&databaseConfig.Port,
		&databaseConfig.User,
		&databaseConfig.Password,
		&databaseConfig.SSLMode,
		&databaseConfig.ConnectionString,
	)

	// TODO (not_important_for_now) - Maybe parse the connection string and fill connection string or host, port and other stuff 

	if err != nil {
		log.Println("No database records found for user with userId: ", userId)
		return nil, err
	}

	return &databaseConfig, nil
}
