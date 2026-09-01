package tenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Manager struct {
	repo *Repository
	cipher *Cipher
	maxOpen int
	maxIdle int
	maxLifetime time.Duration

	mu sync.Mutex
	pools map[int64]*sql.DB
}

func NewManager(repo *Repository, cipher *Cipher, maxOpen, maxIdle int, maxLifetime time.Duration) *Manager {
	if maxOpen < 1 { maxOpen = 10 }
	if maxIdle < 1 { maxIdle = 5 }
	return &Manager{repo: repo, cipher: cipher, maxOpen: maxOpen, maxIdle: maxIdle, maxLifetime: maxLifetime, pools: make(map[int64]*sql.DB)}
}

func (m *Manager) Credentials(ctx context.Context, tenantID int64) (Credentials, error) {
	record, err := m.repo.Get(ctx, tenantID)
	if err != nil { return Credentials{}, err }
	if record.Status != StatusVerified && record.Status != StatusReady {
		return Credentials{}, fmt.Errorf("tenantdb: tenant database is not ready")
	}
	password, err := m.cipher.Decrypt(record.EncryptedPassword)
	if err != nil { return Credentials{}, err }
	return Credentials{Host: record.Host, Port: record.Port, Database: record.DatabaseName, Username: record.Username, Password: password}, nil
}

func (m *Manager) DB(ctx context.Context, tenantID int64) (*sql.DB, error) {
	m.mu.Lock()
	if db := m.pools[tenantID]; db != nil {
		m.mu.Unlock()
		return db, nil
	}
	m.mu.Unlock()

	creds, err := m.Credentials(ctx, tenantID)
	if err != nil { return nil, err }
	db, err := sql.Open("mysql", mysqlDSN(creds))
	if err != nil { return nil, err }
	db.SetMaxOpenConns(m.maxOpen)
	db.SetMaxIdleConns(m.maxIdle)
	db.SetConnMaxLifetime(m.maxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("tenantdb: ping: %w", err)
	}

	m.mu.Lock()
	if existing := m.pools[tenantID]; existing != nil {
		m.mu.Unlock()
		_ = db.Close()
		return existing, nil
	}
	m.pools[tenantID] = db
	m.mu.Unlock()
	return db, nil
}

func (m *Manager) Invalidate(tenantID int64) {
	m.mu.Lock()
	db := m.pools[tenantID]
	delete(m.pools, tenantID)
	m.mu.Unlock()
	if db != nil { _ = db.Close() }
}

func (m *Manager) Close() error {
	m.mu.Lock()
	pools := m.pools
	m.pools = make(map[int64]*sql.DB)
	m.mu.Unlock()
	var first error
	for _, db := range pools {
		if err := db.Close(); err != nil && first == nil { first = err }
	}
	return first
}

func mysqlDSN(c Credentials) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC", c.Username, c.Password, c.Host, c.Port, c.Database)
}
