package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gatepass/internal/config"
	"gatepass/internal/database"
)

const migrationLockName = "passnow_schema_migrations"

var migrationFilePattern = regexp.MustCompile(`^[0-9]+_.+\.up\.sql$`)

type migration struct {
	Name     string
	Path     string
	SQL      string
	Checksum string
}

func main() {
	action := flag.String("action", "up", "up or status")
	dir := flag.String("dir", "migrations", "migration directory")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := ensureMigrationsTable(ctx, db); err != nil {
		log.Fatalf("migrations table: %v", err)
	}

	switch *action {
	case "up":
		if err := runUp(ctx, db, *dir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
	case "status":
		if err := printStatus(ctx, db, *dir); err != nil {
			log.Fatalf("migrate status: %v", err)
		}
	default:
		log.Fatalf("unknown action %q (use up or status)", *action)
	}
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name VARCHAR(255) NOT NULL PRIMARY KEY,
			checksum CHAR(64) NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}

func runUp(ctx context.Context, db *sql.DB, dir string) error {
	if err := acquireLock(ctx, db); err != nil {
		return err
	}
	defer releaseLock(db)

	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		var checksum string
		err := db.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE name = ?", m.Name).Scan(&checksum)
		switch {
		case err == nil:
			if checksum != m.Checksum {
				return fmt.Errorf("%s was already applied but its checksum changed; never edit an applied migration", m.Name)
			}
			log.Printf("skip %s (already applied)", m.Name)
			continue
		case err != sql.ErrNoRows:
			return err
		}

		log.Printf("apply %s", m.Name)
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		for _, statement := range splitSQL(m.SQL) {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("%s: %w", m.Name, err)
			}
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (name, checksum, applied_at) VALUES (?, ?, NOW())",
			m.Name, m.Checksum,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%s: record migration: %w", m.Name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("%s: commit: %w", m.Name, err)
		}
		log.Printf("applied %s", m.Name)
	}
	return nil
}

func printStatus(ctx context.Context, db *sql.DB, dir string) error {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}

	applied := map[string]string{}
	rows, err := db.QueryContext(ctx, "SELECT name, checksum FROM schema_migrations ORDER BY name")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name, checksum string
		if err := rows.Scan(&name, &checksum); err != nil {
			return err
		}
		applied[name] = checksum
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, m := range migrations {
		state := "pending"
		if checksum, ok := applied[m.Name]; ok {
			state = "applied"
			if checksum != m.Checksum {
				state = "checksum-mismatch"
			}
		}
		fmt.Printf("%-20s %s\n", state, m.Name)
	}
	return nil
}

func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !migrationFilePattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(body)
		migrations = append(migrations, migration{
			Name:     entry.Name(),
			Path:     path,
			SQL:      string(body),
			Checksum: fmt.Sprintf("%x", sum[:]),
		})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Name < migrations[j].Name })
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no *.up.sql migrations found in %s", dir)
	}
	return migrations, nil
}

func acquireLock(ctx context.Context, db *sql.DB) error {
	var acquired int
	if err := db.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", migrationLockName).Scan(&acquired); err != nil {
		return err
	}
	if acquired != 1 {
		return fmt.Errorf("could not acquire migration lock")
	}
	return nil
}

func releaseLock(db *sql.DB) {
	_, _ = db.Exec("SELECT RELEASE_LOCK(?)", migrationLockName)
}

// splitSQL separates normal MySQL statements without requiring the driver DSN
// to enable multiStatements. It deliberately understands quotes and comments;
// migrations should still avoid stored procedures/DELIMITER blocks.
func splitSQL(input string) []string {
	var out []string
	var b strings.Builder
	var quote rune
	inLineComment := false
	inBlockComment := false

	runes := []rune(input)
	for i, r := range runes {
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		if inLineComment {
			b.WriteRune(r)
			if r == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			b.WriteRune(r)
			if r == '*' && next == '/' {
				continue
			}
			if r == '/' && i > 0 && runes[i-1] == '*' {
				inBlockComment = false
			}
			continue
		}

		if quote == 0 {
			if r == '-' && next == '-' {
				inLineComment = true
				b.WriteRune(r)
				continue
			}
			if r == '#' {
				inLineComment = true
				b.WriteRune(r)
				continue
			}
			if r == '/' && next == '*' {
				inBlockComment = true
				b.WriteRune(r)
				continue
			}
			if r == '\'' || r == '"' || r == '`' {
				quote = r
				b.WriteRune(r)
				continue
			}
			if r == ';' {
				out = append(out, b.String())
				b.Reset()
				continue
			}
			b.WriteRune(r)
			continue
		}

		b.WriteRune(r)
		if r == quote && (i == 0 || runes[i-1] != '\\') {
			quote = 0
		}
	}
	if strings.TrimSpace(b.String()) != "" {
		out = append(out, b.String())
	}
	return out
}
