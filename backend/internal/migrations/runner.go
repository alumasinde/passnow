package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const LockPrefix = "passnow_schema_migrations"

// Migrations numbered below this predate the uniqueness rule (several share a
// number but have distinct full names, which is what schema_migrations keys on).
// Do NOT rename them: that would make every existing tenant re-run them.
const firstUniqueVersion = 24

var (
	filePattern = regexp.MustCompile(`^([0-9]+)_.+\.up\.sql$`)
	ddlPattern  = regexp.MustCompile(`(?im)^\s*(CREATE|ALTER|DROP|TRUNCATE|RENAME)\b`)
)

type Migration struct {
	Name     string
	Path     string
	SQL      string
	Checksum string
	Version  int
}

func EnsureTable(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (name VARCHAR(255) NOT NULL PRIMARY KEY, checksum CHAR(64) NOT NULL, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations_dirty (name VARCHAR(255) NOT NULL PRIMARY KEY, started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, error TEXT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	return err
}

func RunUp(ctx context.Context, db *sql.DB, dir, lockName string) error {
	if lockName == "" {
		lockName = LockPrefix
	}
	if err := EnsureTable(ctx, db); err != nil {
		return err
	}

	// GET_LOCK is bound to ONE connection. Pin a connection so the lock, every
	// migration and the release all happen on the same session.
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := acquireLock(ctx, conn, lockName); err != nil {
		return err
	}
	defer releaseLock(conn, lockName)

	items, err := Load(dir)
	if err != nil {
		return err
	}
	for _, m := range items {
		var checksum string
		err := conn.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE name = ?", m.Name).Scan(&checksum)
		switch {
		case err == nil:
			if checksum != m.Checksum {
				return fmt.Errorf("%s was already applied but its checksum changed; never edit an applied migration", m.Name)
			}
			continue
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}
		if err := applyOne(ctx, conn, m); err != nil {
			return err
		}
	}
	return nil
}

// applyOne runs a single migration. MySQL commits implicitly on DDL, so a
// failure part-way through a file cannot be rolled back. We therefore record
// the migration as "dirty" BEFORE running it, and clear that only on success.
// A retry then fails fast with a clear message instead of "table already exists".
func applyOne(ctx context.Context, conn *sql.Conn, m Migration) error {
	var started string
	err := conn.QueryRowContext(ctx, "SELECT CAST(started_at AS CHAR) FROM schema_migrations_dirty WHERE name = ?", m.Name).Scan(&started)
	if err == nil {
		return fmt.Errorf("%s started at %s and did not finish; inspect the schema, repair it by hand, then call migrations.ClearDirty", m.Name, started)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO schema_migrations_dirty (name) VALUES (?)", m.Name); err != nil {
		return err
	}

	fail := func(cause error) error {
		if ddlPattern.MatchString(m.SQL) {
			msg := cause.Error()
			if len(msg) > 2000 {
				msg = msg[:2000]
			}
			_, _ = conn.ExecContext(ctx, "UPDATE schema_migrations_dirty SET error = ? WHERE name = ?", msg, m.Name)
		} else {
			// Pure DML: the rollback was complete, so nothing is half-applied.
			_, _ = conn.ExecContext(ctx, "DELETE FROM schema_migrations_dirty WHERE name = ?", m.Name)
		}
		return fmt.Errorf("%s: %w", m.Name, cause)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	for _, statement := range SplitSQL(m.SQL) {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			_ = tx.Rollback()
			return fail(err)
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name, checksum, applied_at) VALUES (?, ?, NOW())", m.Name, m.Checksum); err != nil {
		_ = tx.Rollback()
		return fail(err)
	}
	if err := tx.Commit(); err != nil {
		return fail(err)
	}
	_, err = conn.ExecContext(ctx, "DELETE FROM schema_migrations_dirty WHERE name = ?", m.Name)
	return err
}

// ClearDirty is called by an operator after repairing a half-applied migration
// by hand (either finish it and insert its schema_migrations row, or undo it).
func ClearDirty(ctx context.Context, db *sql.DB, name string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM schema_migrations_dirty WHERE name = ?", name)
	return err
}

func Status(ctx context.Context, db *sql.DB, dir string) ([]string, error) {
	if err := EnsureTable(ctx, db); err != nil {
		return nil, err
	}
	items, err := Load(dir)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "SELECT name, checksum FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := map[string]string{}
	for rows.Next() {
		var n, c string
		if err := rows.Scan(&n, &c); err != nil {
			return nil, err
		}
		applied[n] = c
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	dirty := map[string]bool{}
	drows, err := db.QueryContext(ctx, "SELECT name FROM schema_migrations_dirty")
	if err != nil {
		return nil, err
	}
	defer drows.Close()
	for drows.Next() {
		var n string
		if err := drows.Scan(&n); err != nil {
			return nil, err
		}
		dirty[n] = true
	}
	out := make([]string, 0, len(items))
	for _, m := range items {
		state := "pending"
		if checksum, ok := applied[m.Name]; ok {
			state = "applied"
			if checksum != m.Checksum {
				state = "checksum-mismatch"
			}
		} else if dirty[m.Name] {
			state = "DIRTY"
		}
		out = append(out, fmt.Sprintf("%-20s %s", state, m.Name))
	}
	return out, nil
}

func Load(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	items := make([]Migration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := filePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		version, _ := strconv.Atoi(match[1])
		path := filepath.Join(dir, entry.Name())
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(body)
		items = append(items, Migration{Name: entry.Name(), Path: path, SQL: string(body), Checksum: fmt.Sprintf("%x", sum[:]), Version: version})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	if len(items) == 0 {
		return nil, fmt.Errorf("no *.up.sql migrations found in %s", dir)
	}

	seen := map[int]string{}
	for _, m := range items {
		if m.Version < firstUniqueVersion {
			continue
		}
		if prev, dup := seen[m.Version]; dup {
			return nil, fmt.Errorf("migrations %s and %s share version %04d; versions from %04d on must be unique", prev, m.Name, m.Version, firstUniqueVersion)
		}
		seen[m.Version] = m.Name
	}
	return items, nil
}

func acquireLock(ctx context.Context, conn *sql.Conn, name string) error {
	var got sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", name).Scan(&got); err != nil {
		return err
	}
	if !got.Valid || got.Int64 != 1 {
		return fmt.Errorf("could not acquire migration lock %q", name)
	}
	return nil
}

func releaseLock(conn *sql.Conn, name string) {
	var released sql.NullInt64
	_ = conn.QueryRowContext(context.Background(), "SELECT RELEASE_LOCK(?)", name).Scan(&released)
}
func SplitSQL(input string) []string {
	var out []string; var b strings.Builder; var quote rune; inLine:=false; inBlock:=false
	runes:=[]rune(input)
	for i,r:=range runes {
		next:=rune(0);if i+1<len(runes){next=runes[i+1]}
		if inLine { b.WriteRune(r);if r=='\n'{inLine=false};continue }
		if inBlock { b.WriteRune(r);if r=='/'&&i>0&&runes[i-1]=='*'{inBlock=false};continue }
		if quote==0 {
			if r=='-'&&next=='-' {inLine=true;b.WriteRune(r);continue}
			if r=='#' {inLine=true;b.WriteRune(r);continue}
			if r=='/'&&next=='*' {inBlock=true;b.WriteRune(r);continue}
			if r=='\''||r=='"'||r=='\x60' {quote=r;b.WriteRune(r);continue}
			if r==';' {out=append(out,b.String());b.Reset();continue}
			b.WriteRune(r);continue
		}
		b.WriteRune(r);if r==quote&&(i==0||runes[i-1]!='\\'){quote=0}
	}
	if strings.TrimSpace(b.String())!=""{out=append(out,b.String())}
	return out
}
