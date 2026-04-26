package config

import (
	"context"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	PostgresString   string `env:"POSTGRES_DBSTRING" env-required:"true"`
	S3Endpoint       string `env:"S3_ENDPOINT" env-required:"true"`
	S3AccessKey      string `env:"S3_ACCESS_KEY" env-required:"true"`
	S3SecretKey      string `env:"S3_SECRET_KEY" env-required:"true"`
	RabbitMQHost     string `env:"RABBIT_HOST" env-required:"true"`
	RabbitMQPort     int    `env:"RABBIT_PORT" env-default:"5672"`
	RabbitMQLogin    string `env:"RABBIT_USER" env-required:"true"`
	RabbitMQPassword string `env:"RABBIT_PASSWORD" env-required:"true"`
	ListenAddr       string `env:"LISTEN_ADDR" env-default:"0.0.0.0:4000"`
	JWTSecret        string `env:"JWT_SECRET"`
	LTIIssuer        string `env:"LTI_ISSUER"`
	LTIAuthURL       string `env:"LTI_AUTH_URL"`
	LTITokenURL      string `env:"LTI_TOKEN_URL"`
	LTIJWKSURL       string `env:"LTI_JWKS_URL"`
	LTIClientID      string `env:"LTI_CLIENT_ID"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if _, err := os.Stat(".env"); err == nil {
		return cfg, cleanenv.ReadConfig(".env", cfg)
	} else {
		return cfg, cleanenv.ReadEnv(cfg)
	}
}

func (c *Config) NewPgxPool(ctx context.Context) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(ctx, c.PostgresString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return pool, nil
}
