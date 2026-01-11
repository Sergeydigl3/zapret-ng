package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/twitchtv/twirp"
	"github.com/Sergeydigl3/zapret-discord-youtube-ng/rpc/daemon"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get strategy runner status",
	Long:  `Get the current status of the strategy runner.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetStatus(ctx, &daemon.StatusRequest{})
	if err != nil {
		// Handle Twirp errors with more context
		if twerr, ok := err.(twirp.Error); ok {
			return fmt.Errorf("get status failed: %s (code: %s)", twerr.Msg(), twerr.Code())
		}
		return fmt.Errorf("get status failed: %w", err)
	}

	// Print status
	runningStr := "❌ not running"
	if resp.Running {
		runningStr = "✓ running"
	}

	fmt.Printf("Status:             %s\n", runningStr)
	fmt.Printf("Strategy File:      %s\n", resp.StrategyFile)
	fmt.Printf("Active Queues:      %d\n", resp.ActiveQueues)
	fmt.Printf("Active Processes:   %d\n", resp.ActiveProcesses)
	fmt.Printf("Firewall Backend:   %s\n", resp.FirewallBackend)

	return nil
}
