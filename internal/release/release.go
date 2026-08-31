package release

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Dannie226/cd_watcher/internal/command"
)

var HealthCheckErr error = fmt.Errorf("Health check failed")
var RestartErr error = fmt.Errorf("Restart Failed")

func SetupRelease(releaseDir, releaseName string, restartScript, healthCheckScript string) error {
	linkDir := filepath.Join(releaseDir, "current")
	tmpLinkDir := filepath.Join(releaseDir, "current_tmp")
	relDir := filepath.Join(".", releaseName)

	err := os.Remove(tmpLinkDir)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Failed to remove temporary current symlink")
	}

	err = os.Symlink(relDir, tmpLinkDir)

	if err != nil {
		return fmt.Errorf("Failed to add new symlink: %w", err)
	}

	err = os.Rename(tmpLinkDir, linkDir)

	if err != nil {
		return fmt.Errorf("Failed to rename symlink: %w", err)
	}

	if err := command.RunCommand(
		"/usr/bin/bash",
		"",
		"Restart Script",
		restartScript,
	); err != nil {
		return fmt.Errorf("%w: %w", RestartErr, err)
	}

	if healthCheckScript != "" {
		if err := command.RunCommand(
			"/usr/bin/bash",
			"",
			"Health Check Script",
			healthCheckScript,
		); err != nil {
			return fmt.Errorf("%w: %w", HealthCheckErr, err)
		}
	}

	return nil
}
