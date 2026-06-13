package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Env != "local" {
		t.Fatalf("expected default app env local, got %s", cfg.App.Env)
	}
	if cfg.App.Port != 8080 {
		t.Fatalf("expected default app port 8080, got %d", cfg.App.Port)
	}
	if cfg.Storage.Driver != "memory" {
		t.Fatalf("expected default storage driver memory, got %s", cfg.Storage.Driver)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected default database port 5432, got %d", cfg.Database.Port)
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_HOST", "127.0.0.1")
	t.Setenv("APP_PORT", "9000")
	t.Setenv("STORAGE_DRIVER", "postgres")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("DB_PORT", "5432")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Env != "test" {
		t.Fatalf("expected app env test, got %s", cfg.App.Env)
	}
	if cfg.App.Address() != "127.0.0.1:9000" {
		t.Fatalf("expected address 127.0.0.1:9000, got %s", cfg.App.Address())
	}
	if cfg.Storage.Driver != "postgres" {
		t.Fatalf("expected storage driver postgres, got %s", cfg.Storage.Driver)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected database port 5432, got %d", cfg.Database.Port)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid APP_PORT error")
	}
}

func TestLoadRejectsOutOfRangePort(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected out of range APP_PORT error")
	}
}

func TestDatabaseConnectionStringUsesURL(t *testing.T) {
	cfg := DatabaseConfig{
		URL: "postgres://user:pass@localhost:5432/db?sslmode=disable",
	}

	if cfg.ConnectionString() != cfg.URL {
		t.Fatalf("expected database url, got %s", cfg.ConnectionString())
	}
}

func TestDatabaseConnectionStringBuildsFromParts(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "fitness",
		Password: "fitness",
		Name:     "fitness_tracker",
		SSLMode:  "disable",
	}

	expected := "postgres://fitness:fitness@localhost:5432/fitness_tracker?sslmode=disable"
	if cfg.ConnectionString() != expected {
		t.Fatalf("expected %s, got %s", expected, cfg.ConnectionString())
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	names := []string{
		"APP_ENV",
		"APP_HOST",
		"APP_PORT",
		"STORAGE_DRIVER",
		"DATABASE_URL",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SSLMODE",
	}

	for _, name := range names {
		t.Setenv(name, "")
	}
}
