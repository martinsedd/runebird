package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	SMTP      SMTPConfig      `yaml:"smtp"`
	Templates TemplatesConfig `yaml:"templates"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type SMTPConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"from_address"`
}

type TemplatesConfig struct {
	Path string `yaml:"path"`
}

type RateLimitConfig struct {
	PerHour int `yaml:"per_hour"`
	Burst   int `yaml:"burst"`
}

type LoggingConfig struct {
	FilePath string `yaml:"file_path"`
	Level    string `yaml:"level"`
}

func (c *Config) setDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.SMTP.Host == "" {
		c.SMTP.Host = "localhost"
	}
	if c.SMTP.Port == 0 {
		c.SMTP.Port = 587
	}
	if c.SMTP.FromAddress == "" {
		c.SMTP.FromAddress = "no-reply@runebird.app"
	}

	if c.Templates.Path == "" {
		c.Templates.Path = "./templates"
	}

	if c.RateLimit.PerHour == 0 {
		c.RateLimit.PerHour = 100
	}
	if c.RateLimit.Burst == 0 {
		c.RateLimit.Burst = 5
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.FilePath == "" {
		c.Logging.FilePath = "./logs/runebird.log"
	}
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}

	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.SMTP.Port < 1 || c.SMTP.Port > 65535 {
		return fmt.Errorf("SMTP port must be between 1 and 65535, got %d", c.SMTP.Port)
	}
	if c.SMTP.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if c.SMTP.FromAddress == "" {
		return fmt.Errorf("SMTP from address is required")
	}

	if c.Templates.Path == "" {
		return fmt.Errorf("templates path is required")
	}

	if c.RateLimit.PerHour < 1 {
		return fmt.Errorf("rate limit per hour must be greater than 0, got %d", c.RateLimit.PerHour)
	}
	if c.RateLimit.Burst < 1 {
		return fmt.Errorf("rate limit burst must be greater than 0, got %d", c.RateLimit.Burst)
	}

	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("logging level must be one of debug, info, warn, error; got %s", c.Logging.Level)
	}

	return nil

}

func Load() (*Config, error) {
	path := os.Getenv("EMAILER_CONFIG_PATH")
	if path == "" {
		path = "emailer.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	cfg.setDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &cfg, nil
}
