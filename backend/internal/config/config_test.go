package config

import "testing"

func TestLoadDatabaseRequiresCredentials(t *testing.T) {
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	if _, err := LoadDatabase(); err == nil {
		t.Fatal("expected missing database credentials to fail")
	}
}

func TestLoadDatabaseDoesNotRequireApplicationSecrets(t *testing.T) {
	t.Setenv("DB_USER", "passnow")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "passnow")
	t.Setenv("JWT_SECRET", "")

	cfg, err := LoadDatabase()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBUser != "passnow" || cfg.DBName != "passnow" {
		t.Fatalf("unexpected database config: %+v", cfg)
	}
}

func TestMigrationDSNEnablesMultiStatementsOnlyForMigrations(t *testing.T) {
	cfg := &Config{DBHost: "127.0.0.1", DBPort: "3306", DBUser: "u", DBPassword: "p", DBName: "passnow"}
	if got := cfg.MySQLDSN(); contains(got, "multiStatements=true") {
		t.Fatal("normal application DSN must not enable multiStatements")
	}
	if got := cfg.MySQLMigrationDSN(); !contains(got, "multiStatements=true") {
		t.Fatal("migration DSN must enable multiStatements")
	}
}

func contains(s, want string) bool {
	for i := 0; i+len(want) <= len(s); i++ {
		if s[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
