package server

import (
	"log"
	"sync"
	"time"
)

type bundleData struct {
	transactions []string
	createdAt    time.Time
}

type RpcEndPointServer struct {
	bundles map[string]*bundleData
	mutex   sync.RWMutex
	maxAge  time.Duration
	maxTxs  int
	logger  *log.Logger
}

func NewRpcEndPointServer() *RpcEndPointServer {
	s := &RpcEndPointServer{
		bundles: make(map[string]*bundleData),
		maxAge:  10 * time.Minute,
		maxTxs:  15,
		logger:  log.Default(),
	}

	go s.cleanupLoop()
	return s
}

func (s *RpcEndPointServer) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.cleanup()
	}
}

func (s *RpcEndPointServer) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for id, bundle := range s.bundles {
		if now.Sub(bundle.createdAt) > s.maxAge {
			delete(s.bundles, id)
		}
	}
}

func (s *RpcEndPointServer) getBundleTxs(bundleID string) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if bundle, exists := s.bundles[bundleID]; exists {
		return bundle.transactions, nil
	}
	return []string{}, nil
}

func (s *RpcEndPointServer) addTxToBundle(bundleID, rawTx string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bundle, exists := s.bundles[bundleID]
	if !exists {
		bundle = &bundleData{
			transactions: make([]string, 0),
			createdAt:    time.Now(),
		}
		s.bundles[bundleID] = bundle
	}

	// Check for duplicate tx
	for _, tx := range bundle.transactions {
		if tx == rawTx {
			return nil
		}
	}

	// Add new transaction
	bundle.transactions = append(bundle.transactions, rawTx)

	// Trim if exceeds max transactions
	if len(bundle.transactions) > s.maxTxs {
		bundle.transactions = bundle.transactions[len(bundle.transactions)-s.maxTxs:]
	}

	return nil
}

func (s *RpcEndPointServer) deleteBundleTxs(bundleID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.bundles, bundleID)
	return nil
}
