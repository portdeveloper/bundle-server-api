package server

import (
	"encoding/json"
	"net/http"
)

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
}

func (s *RpcEndPointServer) HandleBundleRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCorsHeaders(w)

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	bundleID := r.URL.Query().Get("id")
	if bundleID == "" {
		http.Error(w, "no bundle id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetBundle(w, bundleID)
	case http.MethodPost:
		s.handleAddTransaction(w, r, bundleID)
	case http.MethodDelete:
		s.handleDeleteBundle(w, bundleID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *RpcEndPointServer) handleGetBundle(w http.ResponseWriter, bundleID string) {
	txs, err := s.getBundleTxs(bundleID)
	if err != nil {
		s.logger.Printf("Error getting bundle txs: %v", err)
		http.Error(w, "failed to get bundle transactions", http.StatusInternalServerError)
		return
	}

	response := BundleResponse{
		BundleID: bundleID,
		RawTxs:   txs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *RpcEndPointServer) handleAddTransaction(w http.ResponseWriter, r *http.Request, bundleID string) {
	var txReq TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if txReq.RawTx == "" {
		http.Error(w, "rawTx is required", http.StatusBadRequest)
		return
	}

	if err := s.addTxToBundle(bundleID, txReq.RawTx); err != nil {
		s.logger.Printf("Error adding tx to bundle: %v", err)
		http.Error(w, "failed to add transaction to bundle", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *RpcEndPointServer) handleDeleteBundle(w http.ResponseWriter, bundleID string) {
	if err := s.deleteBundleTxs(bundleID); err != nil {
		s.logger.Printf("Error deleting bundle: %v", err)
		http.Error(w, "failed to delete bundle", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
