package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *service) CreateUser(user *User) error {
	if user.ID == "" {
		user.ID = utils.GenerateUUID()
	}

	now := time.Now()

	user.CreatedAt = &now
	user.UpdatedAt = &now

	_, err := s.db.Exec(
		`INSERT INTO users 
		id, 
		name, 
		email, 
		email_verified, 
		image, 
		created_at, 
		updated_at
		
		VALUES $1, $2, $3, $4, $5, $6, $7
	`,
		user.ID,
		user.Name,
		user.Email,
		user.EmailVerified,
		user.Image,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (s *service) GetUserByEmail(email string) (*User, error) {
	user := &User{}

	err := s.db.QueryRow(
		`SELECT id, 
		user_name,
		email,
		email_verified,
		image,
		created_at,
		updated_at
		FROM users 
		WHERE email = $1`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) GetUserById(userId string) (*User, error) {
	user := &User{}

	err := s.db.QueryRow(`SELECT 
		id,
		name,
		email,
		email_verified,
		image,
		created_at,
		updated_at
	FROM users WHERE id = $1`, userId).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) CreateAccount(account *Account) error {
	jsonResp, err := json.MarshalIndent(account, "", " ")

	if err != nil {
		log.Fatal("Error occured while formatting json: ", err)
		return err
	}

	fmt.Println("Account: ", string(jsonResp))

	_, err = s.db.Exec(
		`INSERT INTO accounts 
			id,
			user_id,
			provider_type,
			provider_id,
			provider_account_id,
			refresh_token,
			access_token,
			access_token_expires,
			created_at,
			updated_at
		VALUES $1, $2, $3, $4, $5, $6, $7, $8, $9, $10`,
		account.ID,
		account.UserID,
		account.ProviderType,
		account.ProviderID,
		account.ProviderAccountID,
		account.RefreshToken,
		account.AccessToken,
		account.AccessTokenExpires,
		account.CreatedAt,
		account.UpdatedAt,
	)

	return err
}

func (s *service) GetUserByProviderId(providerId string) (*Account, error) {
	account := &Account{}

	err := s.db.QueryRow(`SELECT * FROM accounts WHERE provider_account_id = $1`, providerId).Scan(
		&account.ID,
		&account.UserID,
		&account.ProviderType,
		&account.ProviderID,
		&account.ProviderAccountID,
		&account.RefreshToken,
		&account.AccessToken,
		&account.AccessTokenExpires,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return account, nil
}
