package core

import (
	"errors"
	"sync"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// Balances maps asset symbols to amounts.
type Balances map[string]float64

// Account represents a single user's balance snapshot.
type Account struct {
	Address  string   `json:"address"`
	Balances Balances `json:"balances"`
	Nonce    uint64   `json:"nonce"`
}

// AccountStore is an in‑memory balance store for DevNet.
type AccountStore struct {
	mu       sync.RWMutex
	accounts map[string]*Account
}

// NewAccountStore creates an empty store.
func NewAccountStore() *AccountStore {
	return &AccountStore{
		accounts: make(map[string]*Account),
	}
}

func (s *AccountStore) getOrCreate(addr string) *Account {
	if acc, ok := s.accounts[addr]; ok {
		return acc
	}
	acc := &Account{
		Address:  addr,
		Balances: make(Balances),
	}
	s.accounts[addr] = acc
	return acc
}

func (s *AccountStore) ExpectAndIncrementNonce(addr string, nonce uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.getOrCreate(addr)
	if nonce != acc.Nonce+1 {
		return errors.New("invalid nonce")
	}
	acc.Nonce = nonce
	return nil
}

// GetNonce returns the current nonce for an address.
func (s *AccountStore) GetNonce(addr string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		return acc.Nonce
	}
	return 0
}

// Credit increases a balance for an address/asset pair.
func (s *AccountStore) Credit(addr, asset string, amount float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.getOrCreate(addr)
	acc.Balances[asset] += amount
}

// Debit decreases a balance for an address/asset pair.
func (s *AccountStore) Debit(addr, asset string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc := s.getOrCreate(addr)
	if acc.Balances[asset] < amount {
		return ErrInsufficientFunds
	}
	acc.Balances[asset] -= amount
	return nil
}

// Transfer moves funds between two addresses for a single asset.
func (s *AccountStore) Transfer(from, to, asset string, amount float64) error {
	if err := s.Debit(from, asset, amount); err != nil {
		return err
	}
	s.Credit(to, asset, amount)
	return nil
}

// Snapshot returns a copy of a single account's balances.
func (s *AccountStore) Snapshot(addr string) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		out := &Account{
			Address:  acc.Address,
			Balances: make(Balances),
		}
		for k, v := range acc.Balances {
			out.Balances[k] = v
		}
		return out
	}
	// If we don't know this account yet, still return an empty record.
	return &Account{
		Address:  addr,
		Balances: make(Balances),
	}
}

// SnapshotAll returns a slice of all accounts in the store.
func (s *AccountStore) SnapshotAll() []*Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Account, 0, len(s.accounts))
	for _, acc := range s.accounts {
		copyAcc := &Account{
			Address:  acc.Address,
			Balances: make(Balances),
		}
		for k, v := range acc.Balances {
			copyAcc.Balances[k] = v
		}
		out = append(out, copyAcc)
	}
	return out
}

// SeedDemoBalances initialises some simple demo balances for DevNet.
func SeedDemoBalances(store *AccountStore) {
	store.Credit("treasury", "USD", 1_000_000)
	store.Credit("demo-user", "USD", 10_000)
}
