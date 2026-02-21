package account

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/kiry163/claw-mail-monitor/internal/config"
)

type Manager struct {
	cfg *config.Config
	mu  sync.Mutex
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) List() []config.Account {
	m.mu.Lock()
	defer m.mu.Unlock()

	accounts := make([]config.Account, len(m.cfg.Accounts))
	copy(accounts, m.cfg.Accounts)
	return accounts
}

func (m *Manager) Find(email string) (config.Account, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, account := range m.cfg.Accounts {
		if strings.EqualFold(account.Email, email) {
			return account, true
		}
	}
	return config.Account{}, false
}

func (m *Manager) Add(account config.Account) error {
	if strings.TrimSpace(account.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(account.AuthToken) == "" {
		return errors.New("auth_token is required")
	}

	account.Email = strings.TrimSpace(account.Email)
	account.Provider = strings.TrimSpace(account.Provider)
	if account.Provider == "" {
		return errors.New("provider is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.cfg.Accounts {
		if strings.EqualFold(existing.Email, account.Email) {
			return fmt.Errorf("account already exists: %s", account.Email)
		}
	}

	if !account.Enabled {
		account.Enabled = true
	}

	m.cfg.ApplyDefaults(&account)
	if strings.TrimSpace(account.IMAPHost) == "" {
		return errors.New("imap_host is required")
	}
	if strings.TrimSpace(account.SMTPHost) == "" {
		return errors.New("smtp_host is required")
	}
	m.cfg.Accounts = append(m.cfg.Accounts, account)

	return config.Save(m.cfg)
}

func (m *Manager) Remove(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	idx := -1
	for i, account := range m.cfg.Accounts {
		if strings.EqualFold(account.Email, email) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("account not found: %s", email)
	}

	m.cfg.Accounts = append(m.cfg.Accounts[:idx], m.cfg.Accounts[idx+1:]...)
	return config.Save(m.cfg)
}

func (m *Manager) FirstEnabled() (config.Account, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, account := range m.cfg.Accounts {
		if account.Enabled {
			return account, true
		}
	}
	return config.Account{}, false
}
