package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerURL string
	Port      string

	DatabaseURL string
	RedisURL    string

	S3Bucket    string
	S3Region    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string

	JuiceFSMountPath string
	JuiceFSCacheDir  string
	JuiceFSCacheMB   int
	JuiceFSTrashDays int

	SFTPPort     string
	HostKeyPath  string

	ResendAPIKey string
	FromEmail    string

	StripeSecretKey    string
	StripeWebhookSecret string
}

func Load() *Config {
	return &Config{
		ServerURL: envOr("PIDRIVE_SERVER_URL", "http://localhost:8080"),
		Port:      envOr("PIDRIVE_PORT", "8080"),

		DatabaseURL: envOr("PIDRIVE_DATABASE_URL", "postgres://pidrive:pidrive@localhost:5432/pidrive?sslmode=disable"),
		RedisURL:    envOr("PIDRIVE_REDIS_URL", "redis://localhost:6379/1"),

		S3Bucket:    envOr("PIDRIVE_S3_BUCKET", "pidrive-dev"),
		S3Region:    envOr("PIDRIVE_S3_REGION", "us-east-1"),
		S3Endpoint:  envOr("PIDRIVE_S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey: envOr("PIDRIVE_S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: envOr("PIDRIVE_S3_SECRET_KEY", "minioadmin"),

		JuiceFSMountPath: envOr("PIDRIVE_JUICEFS_MOUNT_PATH", "/mnt/pidrive-master"),
		JuiceFSCacheDir:  envOr("PIDRIVE_JUICEFS_CACHE_DIR", "/tmp/pidrive-cache"),
		JuiceFSCacheMB:   envIntOr("PIDRIVE_JUICEFS_CACHE_SIZE_MB", 5120),
		JuiceFSTrashDays: envIntOr("PIDRIVE_JUICEFS_TRASH_DAYS", 30),

		SFTPPort:    envOr("PIDRIVE_SFTP_PORT", "2022"),
		HostKeyPath: envOr("PIDRIVE_HOST_KEY_PATH", "/var/lib/pidrive/host_key"),

		ResendAPIKey: os.Getenv("PIDRIVE_RESEND_API_KEY"),
		FromEmail:    envOr("PIDRIVE_FROM_EMAIL", "noreply@agents.ressl.ai"),

		StripeSecretKey:    os.Getenv("PIDRIVE_STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("PIDRIVE_STRIPE_WEBHOOK_SECRET"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
