package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Maturin AppConfig `yaml:"maturin"`
}

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	SMTP     SMTPConfig     `yaml:"smtp"`
}

type ServerConfig struct {
	Port    string `yaml:"port"    env:"SERVER_PORT"    env-default:"8080"`
	Timeout int    `yaml:"timeout" env:"SERVER_TIMEOUT" env-default:"30"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"    env:"DB_HOST"     env-default:"localhost"`
	Port     int    `yaml:"port"    env:"DB_PORT"     env-default:"5432"`
	User     string `yaml:"user"    env:"DB_USER"     env-default:"postgres"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `yaml:"name"    env:"DB_NAME"     env-default:"maturin"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE"  env-default:"disable"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"     env:"REDIS_ADDR"     env-default:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
}

type JWTConfig struct {
	Secret          string `env:"JWT_SECRET" env-required:"true"`
	AccessTokenTTL  int    `yaml:"access_token_ttl"  env:"JWT_ACCESS_TTL"  env-default:"15"`    // minutes
	RefreshTokenTTL int    `yaml:"refresh_token_ttl" env:"JWT_REFRESH_TTL" env-default:"10080"` // minutes (7 days)
}

type SMTPConfig struct {
	Host     string `yaml:"host"     env:"SMTP_HOST"     env-default:"smtp.beget.com"`
	Port     int    `yaml:"port"     env:"SMTP_PORT"     env-default:"465"`
	User     string `yaml:"user"     env:"SMTP_USER"     env-default:"noreply@imaginelipa.ru"`
	Password string `env:"SMTP_PASSWORD" env-required:"true"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
