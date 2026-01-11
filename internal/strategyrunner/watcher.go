package strategyrunner

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher watches for changes to the strategy config file.
type ConfigWatcher struct {
	watcher    *fsnotify.Watcher
	configPath string
	onChange   func()
	debounce   time.Duration
	stopCh     chan struct{}
	logger     *slog.Logger
}

// NewConfigWatcher creates a new config watcher.
func NewConfigWatcher(path string, onChange func(), logger *slog.Logger) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	// Watch the config file directory (not the file itself, for better compatibility)
	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	return &ConfigWatcher{
		watcher:    watcher,
		configPath: path,
		onChange:   onChange,
		debounce:   1 * time.Second,
		stopCh:     make(chan struct{}),
		logger:     logger,
	}, nil
}

// Start begins watching for config file changes.
func (cw *ConfigWatcher) Start() error {
	go func() {
		var debounceTimer *time.Timer

		for {
			select {
			case event, ok := <-cw.watcher.Events:
				if !ok {
					return
				}

				// Only care about Write events
				if event.Op&fsnotify.Write == fsnotify.Write {
					cw.logger.Info("config file change detected",
						slog.String("path", event.Name),
						slog.String("op", event.Op.String()),
					)

					// Reset debounce timer
					if debounceTimer != nil {
						debounceTimer.Stop()
					}

					debounceTimer = time.AfterFunc(cw.debounce, func() {
						cw.logger.Info("triggering strategy runner restart due to config change")
						cw.onChange()
					})
				}

			case err, ok := <-cw.watcher.Errors:
				if !ok {
					return
				}
				cw.logger.Error("watcher error", slog.Any("error", err))

			case <-cw.stopCh:
				cw.logger.Info("config watcher stopped")
				return
			}
		}
	}()

	return nil
}

// Stop stops watching for config file changes.
func (cw *ConfigWatcher) Stop() error {
	close(cw.stopCh)
	return cw.watcher.Close()
}
