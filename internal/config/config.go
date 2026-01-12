package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the application configuration.
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Logging        LoggingConfig        `yaml:"logging"`
	StrategyRunner StrategyRunnerConfig `yaml:"strategy_runner"`
}

// ServerConfig contains server-related configuration.
type ServerConfig struct {
	// SocketPath is the path to Unix domain socket.
	// If empty, Unix socket will not be created.
	SocketPath string `yaml:"socket_path" env:"ZAPRET_SOCKET_PATH" env-default:"/run/zapret/zapret-daemon.sock"`

	// NetworkAddress is the network address to listen on (host:port or :port).
	// If empty, network listener will not be created.
	NetworkAddress string `yaml:"network_address" env:"ZAPRET_NETWORK_ADDRESS"`

	// SocketPermissions is the file permissions for Unix socket (octal).
	SocketPermissions os.FileMode `yaml:"socket_permissions" env:"ZAPRET_SOCKET_PERMISSIONS" env-default:"0660"`
}

// LoggingConfig contains logging-related configuration.
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error).
	Level string `yaml:"level" env:"ZAPRET_LOG_LEVEL" env-default:"info"`

	// Format is the log format (json, text).
	Format string `yaml:"format" env:"ZAPRET_LOG_FORMAT" env-default:"text"`
}

// StrategyRunnerConfig contains strategy runner configuration.
type StrategyRunnerConfig struct {
	// Enabled indicates if strategy runner is enabled.
	Enabled bool `yaml:"enabled" env:"ZAPRET_SR_ENABLED" env-default:"false"`

	// ConfigPath is the path to strategy configuration file.
	ConfigPath string `yaml:"config_path" env:"ZAPRET_SR_CONFIG_PATH" env-default:"/etc/zapret/strategy.yaml"`

	// Watch indicates if config file should be watched for changes.
	Watch bool `yaml:"watch" env:"ZAPRET_SR_WATCH" env-default:"true"`

	// NFQWSBinary is the path to nfqws binary.
	NFQWSBinary string `yaml:"nfqws_binary" env:"ZAPRET_SR_NFQWS_BINARY" env-default:"/usr/bin/nfqws"`
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over config file values.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Check if config file exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to access config file: %w", err)
		}
	}

	// Read environment variables (they override file values)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Server.SocketPath == "" && c.Server.NetworkAddress == "" {
		return fmt.Errorf("at least one of socket_path or network_address must be configured")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, text)", c.Logging.Format)
	}

	return nil
}
