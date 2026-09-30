package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// OpenMySQL opens the database configured by PASSNOW_TEST_DSN or TEST_DATABASE_DSN.
// Tests that need MySQL should call this helper and skip cleanly when no DSN is configured.
func OpenMySQL(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PASSNOW_TEST_DSN")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_DSN")
	}
	if dsn == "" {
		t.Skip("MySQL integration test skipped: set PASSNOW_TEST_DSN or TEST_DATABASE_DSN")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open MySQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping MySQL: %v", err)
	}
	return db
}

func RequireTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	var name string
	if err := db.QueryRowContext(context.Background(), `SELECT TABLE_NAME FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`, table).Scan(&name); err != nil {
		t.Fatalf("required table %q unavailable: %v", table, err)
	}
}

func MustExec(t *testing.T, db *sql.DB, query string, args ...any) sql.Result {
	t.Helper()
	res, err := db.ExecContext(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("SQL failed: %v\nquery: %s", err, query)
	}
	return res
}

func MustQueryInt(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("query failed: %v\nquery: %s", err, query)
	}
	return n
}

func AssertRowCount(t *testing.T, db *sql.DB, query string, want int64, args ...any) {
	t.Helper()
	got := MustQueryInt(t, db, fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS q", query), args...)
	if got != want {
		t.Fatalf("row count = %d, want %d", got, want)
	}
}
