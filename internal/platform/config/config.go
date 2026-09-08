package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment, HTTPAddr, DatabaseURL, DefaultLocale   string
	DBPoolMax                                           int32
	SupportedLocales                                    []string
	SessionCookieName                                   string
	SessionTTL                                          time.Duration
	PasswordPepper, WebAuthnRPID                        string
	WebAuthnOrigins                                     []string
	StorageProvider, StorageEndpoint                    string
	StorageRegion, StorageBucket                        string
	StorageAccessKey, StorageSecretKey, StorageBasePath string
	AIBaseURL, AIAPIKey, AIModel                        string
}

func Load() (Config, error) {
	c := Config{
		Environment: env("APP_ENV", "development"), HTTPAddr: env("HTTP_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"),
		DBPoolMax: int32(envInt("DB_POOL_MAX", 10)), DefaultLocale: env("DEFAULT_LOCALE", "zh-CN"), SupportedLocales: split(env("SUPPORTED_LOCALES", "zh-CN,en")),
		SessionCookieName: env("SESSION_COOKIE_NAME", "pcp_session"), SessionTTL: envDuration("SESSION_TTL", 30*24*time.Hour), PasswordPepper: os.Getenv("PASSWORD_PEPPER"),
		WebAuthnRPID: env("WEBAUTHN_RP_ID", "localhost"), WebAuthnOrigins: split(env("WEBAUTHN_RP_ORIGINS", "http://localhost:3001")),
		StorageProvider: env("OBJECT_STORAGE_PROVIDER", "filesystem"), StorageEndpoint: os.Getenv("OBJECT_STORAGE_ENDPOINT"), StorageRegion: env("OBJECT_STORAGE_REGION", "auto"), StorageBucket: env("OBJECT_STORAGE_BUCKET", "pcplatform-dev"), StorageAccessKey: os.Getenv("OBJECT_STORAGE_ACCESS_KEY"), StorageSecretKey: os.Getenv("OBJECT_STORAGE_SECRET_KEY"), StorageBasePath: env("OBJECT_STORAGE_BASE_PATH", "./data/objects"),
		AIBaseURL: strings.TrimRight(os.Getenv("AI_BASE_URL"), "/"), AIAPIKey: os.Getenv("AI_API_KEY"), AIModel: os.Getenv("AI_MODEL"),
	}
	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.SupportedLocales) == 0 {
		return Config{}, fmt.Errorf("SUPPORTED_LOCALES must not be empty")
	}
	return c, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	v, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return v
}
func envDuration(key string, fallback time.Duration) time.Duration {
	v, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil {
		return fallback
	}
	return v
}
func split(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
