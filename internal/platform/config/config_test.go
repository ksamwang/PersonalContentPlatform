package config

import (
	"os"
	"testing"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
}

func TestLoadReadsDatabaseURLFile(t *testing.T) {
	path := t.TempDir() + "/database-url"
	if err := os.WriteFile(path, []byte("postgres://from-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", path)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://from-secret" {
		t.Fatalf("unexpected database URL %q", cfg.DatabaseURL)
	}
}

func TestLoadDotEnvFromConfiguredPath(t *testing.T) {
	path := t.TempDir() + "/custom.env"
	if err := os.WriteFile(path, []byte("PCP_TEST_DOTENV_VALUE=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	previous, existed := os.LookupEnv("PCP_TEST_DOTENV_VALUE")
	if err := os.Unsetenv("PCP_TEST_DOTENV_VALUE"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv("PCP_TEST_DOTENV_VALUE", previous)
		} else {
			_ = os.Unsetenv("PCP_TEST_DOTENV_VALUE")
		}
	})
	t.Setenv("APP_ENV_FILE", path)
	if err := loadDotEnv(); err != nil {
		t.Fatal(err)
	}
	if value := os.Getenv("PCP_TEST_DOTENV_VALUE"); value != "loaded" {
		t.Fatalf("unexpected dotenv value %q", value)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SUPPORTED_LOCALES", "zh-CN,en")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultLocale != "zh-CN" || len(cfg.SupportedLocales) != 2 {
		t.Fatalf("unexpected locale config: %#v", cfg)
	}
}
