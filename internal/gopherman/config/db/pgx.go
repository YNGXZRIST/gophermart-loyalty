package db

type Config struct {
	DNS        string
	MaxRetries int
}

const MaxRetries int = 3

func NewCfg(dns string) *Config {
	return &Config{
		DNS:        dns,
		MaxRetries: MaxRetries,
	}
}
