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
	MinIOHost        string `env:"MINIO_HOST" env-required:"true"`
	MinIOLogin       string `env:"MINIO_LOGIN" env-required:"true"`
	MinIOPassword    string `env:"MINIO_PASSWORD" env-required:"true"`
	RabbitMQHost     string `env:"RABBIT_HOST" env-required:"true"`
	RabbitMQPort     int    `env:"RABBIT_PORT" env-default:"5672"`
	RabbitMQLogin    string `env:"RABBIT_USER" env-required:"true"`
	RabbitMQPassword string `env:"RABBIT_PASSWORD" env-required:"true"`
	ListenAddr       string `env:"LISTEN_ADDR" env-default:"4000"`
	JWTSecret        string `env:"JWT_SECRET" env-required:"true"`
	LTIIssuer        string `env:"LTI_ISSUER" env-required:"false"`
	LTIAuthURL       string `env:"LTI_AUTH_URL" env-required:"false"`
	LTITokenURL      string `env:"LTI_TOKEN_URL" env-required:"false"`
	LTIJWKSURL       string `env:"LTI_JWKS_URL" env-required:"false"`
	LTIClientID      string `env:"LTI_CLIENT_ID" env-required:"false"`
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
