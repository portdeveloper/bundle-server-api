package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleBundleRequest(t *testing.T) {
	s := NewRpcEndPointServer()

	// Test cases
	tests := []struct {
		name         string
		method       string
		bundleID     string
		body         string
		expectedCode int
		expectedTxs  []string
	}{
		{
			name:         "Get Empty Bundle",
			method:       http.MethodGet,
			bundleID:     "test-bundle-1",
			expectedCode: http.StatusOK,
			expectedTxs:  []string{},
		},
		{
			name:         "Add Transaction",
			method:       http.MethodPost,
			bundleID:     "test-bundle-1",
			body:         `{"rawTx":"0x123"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Bundle With Transaction",
			method:       http.MethodGet,
			bundleID:     "test-bundle-1",
			expectedCode: http.StatusOK,
			expectedTxs:  []string{"0x123"},
		},
		{
			name:         "Delete Bundle",
			method:       http.MethodDelete,
			bundleID:     "test-bundle-1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Deleted Bundle",
			method:       http.MethodGet,
			bundleID:     "test-bundle-1",
			expectedCode: http.StatusOK,
			expectedTxs:  []string{},
		},
		{
			name:         "Missing Bundle ID",
			method:       http.MethodGet,
			bundleID:     "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.method == http.MethodPost && tt.body != "" {
				req = httptest.NewRequest(tt.method, "/bundle?id="+tt.bundleID, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/bundle?id="+tt.bundleID, nil)
			}

			rr := httptest.NewRecorder()
			s.HandleBundleRequest(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedCode)
			}

			if tt.method == http.MethodGet && tt.expectedCode == http.StatusOK {
				var response BundleResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(response.RawTxs) != len(tt.expectedTxs) {
					t.Errorf("handler returned wrong number of transactions: got %v want %v",
						len(response.RawTxs), len(tt.expectedTxs))
				}

				for i, tx := range response.RawTxs {
					if tx != tt.expectedTxs[i] {
						t.Errorf("handler returned wrong transaction: got %v want %v", tx, tt.expectedTxs[i])
					}
				}
			}
		})
	}
}

func TestBundleExpiry(t *testing.T) {
	s := NewRpcEndPointServer()
	s.maxAge = 100 * time.Millisecond // Short expiry for testing

	// Add a transaction
	bundleID := "test-expiry"
	err := s.addTxToBundle(bundleID, "0x123")
	if err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}

	// Verify transaction was added
	txs, err := s.getBundleTxs(bundleID)
	if err != nil || len(txs) != 1 || txs[0] != "0x123" {
		t.Errorf("Transaction not added correctly")
	}

	// Wait for expiry
	time.Sleep(200 * time.Millisecond)

	// Force cleanup
	s.cleanup()

	// Verify bundle was removed
	txs, err = s.getBundleTxs(bundleID)
	if err != nil || len(txs) != 0 {
		t.Errorf("Bundle not expired correctly")
	}
}

func TestMaxTransactions(t *testing.T) {
	s := NewRpcEndPointServer()
	s.maxTxs = 2 // Small limit for testing
	bundleID := "test-max"

	// Add three transactions
	transactions := []string{"0x1", "0x2", "0x3"}
	for _, tx := range transactions {
		if err := s.addTxToBundle(bundleID, tx); err != nil {
			t.Fatalf("Failed to add transaction: %v", err)
		}
	}

	// Verify only last two transactions are kept
	txs, err := s.getBundleTxs(bundleID)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(txs) != 2 {
		t.Errorf("Wrong number of transactions: got %v want 2", len(txs))
	}

	expectedTxs := transactions[len(transactions)-2:]
	for i, tx := range txs {
		if tx != expectedTxs[i] {
			t.Errorf("Wrong transaction at position %d: got %v want %v", i, tx, expectedTxs[i])
		}
	}
}
