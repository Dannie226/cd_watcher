package release

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

var HealthCheckErr error = fmt.Errorf("Health check failed")

func SetupRelease(releaseDir, releaseName string, restartScript, healthCheckScript string) error {
	linkDir := filepath.Join(releaseDir, "current")
	relDir := filepath.Join(".", releaseName)

	err := os.Remove(linkDir)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Failed to remove old symlink: %w", err)
	}

	err = os.Symlink(relDir, linkDir)

	if err != nil {
		return fmt.Errorf("Failed to add new symlink: %w", err)
	}

	resCmd := exec.Command("/usr/bin/bash", restartScript)

	if err := resCmd.Run(); err != nil {
		return fmt.Errorf("Failed to restart server: %w", err)
	}

	if healthCheckScript != "" {
		chkCmd := exec.Command("/usr/bin/bash")

		if err := chkCmd.Run(); err != nil {
			return fmt.Errorf("%w: %w", HealthCheckErr, err)
		}
	}

	return nil
}
