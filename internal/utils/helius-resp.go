package utils

type HeliusWebhookResponse struct {
	WebhookId        string    `json:"webhookID"`
	Wallet           string    `json:"wallet"`
	WebhookURL       string    `json:"webhookURL"`
	TransactionTypes []string  `json:"transactionTypes"`
	AccountAddresses []string  `json:"accountAddresses"`
	WebhookType      string    `json:"webhookType"`
	TxnStatus        TxnStatus `json:"txnStatus"`
	AuthHeader       string    `json:"authHeader"`
}

type TxnStatus string

const (
	TxnStatusAll     TxnStatus = "all"
	TxnStatusSuccess TxnStatus = "success"
	TxnStatusFailed  TxnStatus = "failed"
)

type HeliusRpcResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Id      string `json:"id"`
	Result  Result `json:"result"`
}

type Result struct {
	Intf    string  `json:"interface"`
	Id      string  `json:"id"`
	Content Content `json:"content"`
}

type Content struct {
	Schema   string         `json:"$schema"`
	Json_Uri string         `json:"json_uri"`
	Files    []any          `json:"files"`
	Metadata *TokenMetadata `json:"metadata"`
}

type TokenMetadata struct {
	Attributes    []any  `json:"attributes"`
	Description   string `json:"description"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	TokenStandard string `json:"token_standard"`
}
