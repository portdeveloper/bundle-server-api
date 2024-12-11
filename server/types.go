package server

type BundleResponse struct {
	BundleID string   `json:"bundleId"`
	RawTxs   []string `json:"rawTxs"`
}

type TransactionRequest struct {
	RawTx string `json:"rawTx"`
}
