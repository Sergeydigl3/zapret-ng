package strategyrunner

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProcessManager manages nfqws daemon processes.
type ProcessManager struct {
	binaryPath string
	processes  []*os.Process
	logger     *slog.Logger
	mu         sync.Mutex
}

// ProcessConfig contains configuration for a single nfqws process.
type ProcessConfig struct {
	QueueNum int
	Args     []string
}

// NewProcessManager creates a new process manager.
func NewProcessManager(binaryPath string, logger *slog.Logger) *ProcessManager {
	return &ProcessManager{
		binaryPath: binaryPath,
		processes:  []*os.Process{},
		logger:     logger,
	}
}

// Start starts a new nfqws process.
func (pm *ProcessManager) Start(cfg *ProcessConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Build command arguments
	args := []string{
		"--daemon",
		fmt.Sprintf("--qnum=%d", cfg.QueueNum),
	}
	args = append(args, cfg.Args...)

	cmd := exec.Command(pm.binaryPath, args...)

	pm.logger.Info("starting nfqws process",
		slog.Int("queue", cfg.QueueNum),
		slog.String("binary", pm.binaryPath),
		slog.String("args", strings.Join(args, " ")),
	)

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start nfqws: %w", err)
	}

	// Track the process
	pm.processes = append(pm.processes, cmd.Process)

	return nil
}

// StopAll stops all tracked processes gracefully.
func (pm *ProcessManager) StopAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errs []string

	for _, proc := range pm.processes {
		pm.logger.Info("stopping nfqws process", slog.Int("pid", proc.Pid))

		// Send SIGTERM
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			pm.logger.Warn("failed to signal process", slog.Int("pid", proc.Pid), slog.Any("error", err))
			errs = append(errs, fmt.Sprintf("process %d signal failed: %v", proc.Pid, err))
		}

		// Wait with timeout
		done := make(chan error, 1)
		go func() {
			_, err := proc.Wait()
			done <- err
		}()

		// Wait up to 5 seconds for graceful shutdown
		select {
		case <-done:
			pm.logger.Info("nfqws process stopped", slog.Int("pid", proc.Pid))
		case <-time.After(5 * time.Second):
			pm.logger.Warn("process did not stop, killing", slog.Int("pid", proc.Pid))
			if err := proc.Kill(); err != nil {
				pm.logger.Error("failed to kill process", slog.Int("pid", proc.Pid), slog.Any("error", err))
				errs = append(errs, fmt.Sprintf("process %d kill failed: %v", proc.Pid, err))
			}
		}
	}

	pm.processes = nil

	if len(errs) > 0 {
		return fmt.Errorf("process cleanup errors: %v", strings.Join(errs, "; "))
	}

	return nil
}

// Count returns the number of running processes.
func (pm *ProcessManager) Count() int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return len(pm.processes)
}
