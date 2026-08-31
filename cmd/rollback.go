package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/Dannie226/cd_watcher/internal/email"
	"github.com/Dannie226/cd_watcher/internal/queries"
	"github.com/Dannie226/cd_watcher/internal/release"
)

func rollback(cfg *config.Config, client *email.EmailClient, n int) int {
	slog.Info("Starting rollback", "count", n)

	var err error

	if n == 1 {
		err = client.SendEmail(config.RollbackStartEvent, "Starting server rollback")
	} else {
		client.SendEmail(
			config.RollbackStartEvent,
			fmt.Sprintf(
				"Starting server rollback.\n Rolling back %d releases",
				n,
			),
		)
	}

	if err != nil {
		slog.Warn("Failed to send rollback start email")
	}

	versions, err := queries.GetLastVersions(cfg.VersionConn, n+1)

	if err != nil {
		slog.Error("Failed to get rollback versions", "error", err)
		err = client.SendEmail(config.RollbackFinishEvent, "Failed to rollback server")

		if err != nil {
			slog.Warn("Failed to send rollback fail email", "failure", "version", "error", err)
		}

		return 1
	}

	if len(versions) < 2 {
		slog.Error("No version to rollback to", "version count", len(versions))
		err = client.SendEmail(config.RollbackFinishEvent, "Failed to rollback server")

		if err != nil {
			slog.Warn("Failed to send rollback fail email", "failure", "too few versions", "error", err)
		}

		return 1
	}

	if len(versions) < n+1 {
		n = len(versions) - 1
	}

	err = release.SetupRelease(cfg.ReleaseDir, versions[n].FolderName, cfg.ReloadScript, cfg.HealthScript)

	if err != nil {
		slog.Error("Failed to restart server to rolled back version", "error", err)
		err = client.SendEmail(config.RollbackFinishEvent, "Failed to roll back server")

		if err != nil {
			slog.Warn("Failed to send rollback fail email", "failure", "restart", "error", err)
		}

		return 1
	}

	i := 0

	for ; i < n; i++ {
		err = os.RemoveAll(filepath.Join(cfg.ReleaseDir, versions[i].FolderName))

		if err != nil {
			slog.Error("Failed to remove previous releases", "error", err)
			err = client.SendEmail(config.RollbackFinishEvent, "Failed to roll back server")

			if err != nil {
				slog.Warn("Failed to send rollback fail email", "failure", "fs remove", "error", err)
			}

			return 1
		}
	}

	err = queries.RemoveVersions(cfg.VersionConn, n)

	if err != nil {
		slog.Error("Failed to remove releases from database", "error", err)
		err = client.SendEmail(config.RollbackFinishEvent, "Failed to roll back server")

		if err != nil {
			slog.Warn("Failed to send rollback fail email", "failure", "db remove", "error", err)
		}

		return 1
	}

	err = client.SendEmail(config.RollbackFinishEvent, "Successfully Rolled back server")

	if err != nil {
		slog.Warn("Failed to send rollback success email", "error", err)
	}

	return 0
}
