package database

import (
	"time"
)

type User struct {
	ID            string                 `db:"id"`
	Name          *string                `db:"name"`
	Email         *string                `db:"email"`
	EmailVerified bool                   `db:"email_verified"`
	Image         *string                `db:"image"`
	CreatedAt     *time.Time             `db:"created_at"`
	UpdatedAt     *time.Time             `db:"updated_at"`
	Accounts      []Account              `db:"-"`
	Sessions      []Session              `db:"-"`
	DbCredential  UserDatabaseCredential `db:"-"`
	Indexing      []IndexingToken        `db:"-"`
}

type Account struct {
	ID                 string     `db:"id"`
	UserID             string     `db:"user_id"`
	ProviderType       string     `db:"provider_type"`
	ProviderID         string     `db:"provider_id"`
	ProviderAccountID  string     `db:"provider_account_id"`
	RefreshToken       *string    `db:"refresh_token"`
	AccessToken        *string    `db:"access_token"`
	AccessTokenExpires *time.Time `db:"access_token_expires"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type Session struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Expires      time.Time `db:"expires"`
	SessionToken string    `db:"session_token"`
	AccessToken  string    `db:"access_token"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// VerificationRequest represents an email verification request
type VerificationRequest struct {
	ID         string    `db:"id"`
	Identifier string    `db:"identifier"`
	Token      string    `db:"token"`
	Expires    time.Time `db:"expires"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type UserDatabaseCredential struct {
	ID              string     `db:"id"`
	UserId          string     `db:"user_id"`
	DatabaseName    string     `db:"db_name"`
	Host            string     `db:"host"`
	User            string     `db:"user"`
	Port            uint16     `db:"port"`
	Password        string     `db:"password"`
	SSLMode         string     `db:"ssl_mode"`
	ConnectionLimit *int8      `db:"connection_limit"`
	Status          string     `db:"status"`
	LastConnectedAt *time.Time `db:"last_connected_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	ErrorMessage    string     `db:"error_message"`
}

type IndexingToken struct {
	Id              string           `db:"id"`
	UserId          string           `db:"user_id"`
	TokenAddress    string           `db:"token_address"`
	TokenName       string           `db:"token_name"`
	TokenSymbol     string           `db:"token_symbol"`
	Strategy        IndexingStrategy `db:"indexing_strategy"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
	IndexingStopped bool             `db:"indexing_stopped"`
}

type IndexingStrategy string

const (
	NFTCurrentBids           IndexingStrategy = "nft_current_bids"
	NFTCurrentPrices         IndexingStrategy = "nft_current_prices"
	TokensAvailableToBorrow  IndexingStrategy = "tokens_available_to_borrow"
	TokenCrossPlatformPrices IndexingStrategy = "token_cross_platform_prices"
)

var StrategyTransactionTypes = map[IndexingStrategy][]string{
	// 1. Track current active bids on NFTs
	NFTCurrentBids: {
		"NFT_BID",
		"NFT_BID_CANCELLED",
		"NFT_GLOBAL_BID",
		"NFT_GLOBAL_BID_CANCELLED",
		"NFT_AUCTION_CREATED",
		"NFT_AUCTION_UPDATED",
		"NFT_AUCTION_CANCELLED",
		"NFT_SALE", // To remove bids when NFTs are sold
	},

	// 2. Track current listing prices of NFTs
	NFTCurrentPrices: {
		"NFT_LISTING",
		"NFT_CANCEL_LISTING",
		"NFT_SALE",
		"UPDATE_ITEM",
		"LIST_ITEM",
		"DELIST_ITEM",
		"NFT_RENT_LISTING",
		"NFT_RENT_UPDATE_LISTING",
		"NFT_RENT_CANCEL_LISTING",
	},

	// 3. Track available tokens to borrow
	TokensAvailableToBorrow: {
		"LOAN",
		"RESCIND_LOAN",
		"OFFER_LOAN",
		"REPAY_LOAN",
		"TAKE_LOAN",
		"FORECLOSE_LOAN",
		"ADD_TO_POOL",
		"REMOVE_FROM_POOL",
		"DEPOSIT",
		"WITHDRAW",
	},

	// 4. Track token prices across platforms
	TokenCrossPlatformPrices: {
		"SWAP",
		"INIT_SWAP",
		"CANCEL_SWAP",
		"REJECT_SWAP",
		"TOKEN_MINT",
		"TRANSFER",
		"PLATFORM_FEE",
		"FILL_ORDER",
		"UPDATE_ORDER",
		"CREATE_ORDER",
		"CLOSE_ORDER",
		"CANCEL_ORDER",
	},
}
