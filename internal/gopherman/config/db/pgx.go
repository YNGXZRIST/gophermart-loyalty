package db

type Config struct {
	DNS        string
	MaxRetries int
}

func NewCfg(dns string) *Config {
	return &Config{
		DNS:        dns,
		MaxRetries: 3,
	}
}
