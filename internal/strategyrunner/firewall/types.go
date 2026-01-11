package firewall

import (
	"context"
	"fmt"
)

// Firewall is the interface for firewall implementations.
type Firewall interface {
	// Setup prepares the firewall (creates tables/chains)
	Setup(ctx context.Context) error

	// AddRule adds a firewall rule
	AddRule(ctx context.Context, rule *Rule) error

	// RemoveAll removes all rules and cleans up
	RemoveAll(ctx context.Context) error

	// Close closes the firewall connection
	Close() error
}

// Rule represents a firewall rule.
type Rule struct {
	// Protocol is the protocol ("tcp" or "udp")
	Protocol string

	// Ports is a list of ports or port ranges
	Ports []string

	// QueueNum is the NFQUEUE number
	QueueNum int

	// Interface is the network interface ("" for all)
	Interface string

	// Comment is a rule comment
	Comment string
}

// Config contains firewall configuration.
type Config struct {
	// Backend is the firewall backend ("nftables" or "iptables")
	Backend string

	// TableName is the nftables table name
	TableName string

	// ChainName is the chain name
	ChainName string

	// Interface is the network interface
	Interface string
}

// NewFirewall creates a new firewall instance based on the backend.
func NewFirewall(cfg *Config) (Firewall, error) {
	switch cfg.Backend {
	case "nftables":
		return NewNftablesFirewall(cfg)
	case "iptables":
		return NewIptablesFirewall(cfg)
	default:
		return nil, fmt.Errorf("unknown firewall backend: %s", cfg.Backend)
	}
}
