package kafka

import (
	"math/big"
)

type TokenBalanceChange struct {
	Mint           string       `json:"mint" db:"mint"`
	RawTokenAmount RawTokenAmnt `json:"rawTokenAmount" db:"raw_token_amount"`
	TokenAccount   string       `json:"tokenAccount" db:"token_account"`
	UserAccount    string       `json:"userAccount" db:"user_account"`
}

type RawTokenAmnt struct {
	Decimals    int8   `json:"decimals" db:"decimals"`
	TokenAmount string `json:"tokenAmount" db:"token_amount"`
}

type AccountData struct {
	ID                  string               `json:"id" db:"id"`
	Account             string               `json:"account" db:"account"`
	NativeBalanceChange *big.Int             `json:"nativeBalanceChange" db:"native_balance_change"`
	TokenBalanceChanges []TokenBalanceChange `json:"tokenBalanceChanges" db:"token_balance_changes"`
}

type NativeInputOutput struct {
	Account string `json:"account" db:"account"`
	Amount  string `json:"amount" db:"amount"`
}

type InnerSwaps struct {
	NativeFees   []interface{} `json:"nativeFees" db:"native_fees"`
	ProgramInfo  ProgramInf    `json:"programInfo" db:"program_info"`
	TokenFees    []interface{} `json:"tokenFees" db:"token_fees"`
	TokenInputs  []TokenInput  `json:"tokenInputs" db:"token_inputs"`
	TokenOutputs []TokenOutput `json:"tokenOutputs" db:"token_outputs"`
}

type ProgramInf struct {
	Account         string `json:"account" db:"account"`
	InstructionName string `json:"instructionName" db:"instruction_name"`
	ProgramName     string `json:"programName" db:"program_name"`
	Source          string `json:"source" db:"source"`
}

type TokenInput struct {
	FromTokenAccount string  `json:"fromTokenAccount" db:"from_token_account"`
	FromUserAccount  string  `json:"fromUserAccount" db:"from_user_account"`
	Mint             string  `json:"mint" db:"mint"`
	ToTokenAccount   string  `json:"toTokenAccount" db:"to_token_account"`
	ToUserAccount    string  `json:"toUserAccount" db:"to_user_account"`
	TokenAmount      float64 `json:"tokenAmount" db:"token_amount"`
	TokenStandard    string  `json:"tokenStandard" db:"token_standard"`
}

type TokenOutput struct {
	FromTokenAccount string  `json:"fromTokenAccount" db:"from_token_account"`
	FromUserAccount  string  `json:"fromUserAccount" db:"from_user_account"`
	Mint             string  `json:"mint" db:"mint"`
	ToTokenAccount   string  `json:"toTokenAccount" db:"to_token_account"`
	ToUserAccount    string  `json:"toUserAccount" db:"to_user_account"`
	TokenAmount      float64 `json:"tokenAmount" db:"token_amount"`
	TokenStandard    string  `json:"tokenStandard" db:"token_standard"`
}

type EventsSwap struct {
	ID           string            `json:"id" db:"id"`
	Swap         []InnerSwaps      `json:"swap" db:"swap"`
	NativeFees   []interface{}     `json:"nativeFees" db:"native_fees"`
	NativeInput  NativeInputOutput `json:"nativeInput" db:"native_input"`
	NativeOutput NativeInputOutput `json:"nativeOutput" db:"native_output"`
	TokenFees    []interface{}     `json:"tokenFees" db:"token_fees"`
	TokenInputs  []interface{}     `json:"tokenInputs" db:"token_inputs"`
	TokenOutputs []interface{}     `json:"tokenOutputs" db:"token_outputs"`
}

type Instruction struct {
	Id                string             `db:"id"`
	Accounts          []string           `json:"accounts" db:"accounts"`
	Data              string             `json:"data" db:"data"`
	InnerInstructions []InnerInstruction `json:"innerInstructions" db:"inner_instructions"`
}

type InnerInstruction struct {
	Accounts  []string `json:"accounts" db:"accounts"`
	Data      string   `json:"data" db:"data"`
	ProgramId string   `json:"programId" db:"program_id"`
}

type NativeTransfer struct {
	Amount          *big.Int `json:"amount" db:"amount"`
	FromUserAccount string   `json:"fromUserAccount" db:"from_user_account"`
	ToUserAccount   string   `json:"toUserAccount" db:"to_user_account"`
}

type TokenTransfer struct {
	FromTokenAccount string  `json:"fromTokenAccount" db:"from_token_account"`
	FromUserAccount  string  `json:"fromUserAccount" db:"from_user_account"`
	Mint             string  `json:"mint" db:"mint"`
	ToTokenAccount   string  `json:"toTokenAccount" db:"to_token_account"`
	ToUserAccount    string  `json:"toUserAccount" db:"to_user_account"`
	TokenAmount      float64 `json:"tokenAmount" db:"token_amount"`
	TokenStandard    string  `json:"tokenStandard" db:"token_standard"`
}

type WebhookPayload struct {
	AccountData      []AccountData          `json:"accountData" db:"account_data"`
	Description      string                 `json:"description" db:"description"`
	Events           map[string]interface{} `json:"events" db:"events"`
	Fee              int32                  `json:"fee" db:"fee"`
	FeePayer         string                 `json:"feePayer" db:"fee_payer"`
	Instructions     []Instruction          `json:"instructions" db:"instructions"`
	NativeTransfers  []NativeTransfer       `json:"nativeTransfers" db:"native_transfers"`
	Signature        string                 `json:"signature" db:"signature"`
	Slot             int64                  `json:"slot" db:"slot"`
	Source           string                 `json:"source" db:"source"`
	Timestamp        int64                  `json:"timestamp" db:"timestamp"`
	TokenTransfers   []TokenTransfer        `json:"tokenTransfers" db:"token_transfers"`
	TransactionError interface{}            `json:"transactionError" db:"transaction_error"`
	Type             string                 `json:"type" db:"type"`
}

type AddressSet map[string]bool

func (s AddressSet) Add(address string) {
	s[address] = true
}

func (s AddressSet) Contains(address string) bool {
	return s[address]
}