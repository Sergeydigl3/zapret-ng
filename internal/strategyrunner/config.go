package strategyrunner

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the strategy runner configuration.
type Config struct {
	// Interface is the network interface to apply rules to ("eth0", "any", etc.)
	Interface string `yaml:"interface" env:"ZAPRET_INTERFACE" env-default:"any"`

	// GameFilter enables filtering of game ports (1024-65535)
	GameFilter bool `yaml:"gamefilter" env:"ZAPRET_GAMEFILTER" env-default:"true"`

	// GameFilterPorts specifies the port range for game filter
	GameFilterPorts string `yaml:"gamefilter_ports" env:"ZAPRET_GAMEFILTER_PORTS" env-default:"1024-65535"`

	// StrategyFile is the path to the .bat strategy file
	StrategyFile string `yaml:"strategy_file" env:"ZAPRET_STRATEGY_FILE"`

	// Firewall contains firewall backend configuration
	Firewall FirewallConfig `yaml:"firewall"`

	// BinaryPath is the path to nfqws binary (from main config)
	BinaryPath string

	// ConfigPath is the path to this config file (for watcher)
	ConfigPath string

	// Watch indicates if config file should be watched for changes
	Watch bool
}

// FirewallConfig contains firewall backend settings.
type FirewallConfig struct {
	// Backend is the firewall backend to use ("nftables" or "iptables")
	Backend string `yaml:"backend" env:"ZAPRET_FIREWALL_BACKEND" env-default:"nftables"`

	// TableName is the nftables table name (only for nftables backend)
	TableName string `yaml:"table_name" env:"ZAPRET_FIREWALL_TABLE_NAME" env-default:"inet zapretunix"`

	// ChainName is the chain name to use
	ChainName string `yaml:"chain_name" env:"ZAPRET_FIREWALL_CHAIN_NAME" env-default:"output"`
}

// LoadStrategyConfig loads strategy configuration from file and environment variables.
func LoadStrategyConfig(path string) (*Config, error) {
	cfg := &Config{
		Firewall: FirewallConfig{
			Backend:   "nftables",
			TableName: "inet zapretunix",
			ChainName: "output",
		},
	}

	// Check if config file exists
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := cleanenv.ReadConfig(path, cfg); err != nil {
				return nil, fmt.Errorf("failed to read strategy config file: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to access strategy config file: %w", err)
		}
	}

	// Read environment variables (they override file values)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	cfg.ConfigPath = path

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.StrategyFile == "" {
		return fmt.Errorf("strategy_file must be specified")
	}

	if _, err := os.Stat(c.StrategyFile); err != nil {
		return fmt.Errorf("strategy file not found: %s", c.StrategyFile)
	}

	validBackends := map[string]bool{"nftables": true, "iptables": true}
	if !validBackends[c.Firewall.Backend] {
		return fmt.Errorf("invalid firewall backend: %s (must be 'nftables' or 'iptables')", c.Firewall.Backend)
	}

	if c.Interface == "" && c.Interface != "any" {
		return fmt.Errorf("interface must be specified or set to 'any'")
	}

	return nil
}
