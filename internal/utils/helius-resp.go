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
