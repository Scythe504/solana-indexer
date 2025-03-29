package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) CreateDatabaseForUser(userId string, dbCred UserDatabaseCredential) error {
	var connString string

	var jsonBlob []byte
	jsonBlob, _ = json.MarshalIndent(dbCred, "", " ")

	fmt.Println(string(jsonBlob))
	if dbCred.ConnectionString == nil {
		fmt.Println("I came here for some reason")
		connString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require&search_path=public", *dbCred.User, *dbCred.Password, *dbCred.Host, *dbCred.Port, *dbCred.DatabaseName)
	} else {
		fmt.Println("I came here for some reason")
		connString = *dbCred.ConnectionString
		fmt.Printf("\n\n%s\n\n",*dbCred.ConnectionString)
	}

	fmt.Println("db-credentials.go :21", connString)

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
		INSERT INTO user_database_credentials
		(
			id,
			user_id,
			db_name,
			host,
			db_user,
			port,
			db_password, 
			ssl_mode, 
			connection_string,
			connection_limit,
			last_connected_at,
			created_at,
			updated_at, 
			error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, utils.GenerateUUID(),
		userId,
		dbCred.DatabaseName,
		dbCred.Host,
		dbCred.User,
		dbCred.Port,
		dbCred.Password,
		dbCred.SSLMode,
		dbCred.ConnectionString,
		dbCred.ConnectionLimit,
		now,
		now,
		now,
		"",
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
		SELECT id,
			user_id,
			db_name,
			host,
			port,
			user,
			db_password,
			ssl_mode,
			connection_string
		FROM user_database_credentials
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
