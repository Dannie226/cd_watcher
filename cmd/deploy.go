package main

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/Dannie226/cd_watcher/internal/email"
	"github.com/Dannie226/cd_watcher/internal/release"
	"github.com/Dannie226/cd_watcher/internal/unpack"
)

func deploy(cfg *config.Config, client *email.EmailClient) int {
	buf := strings.Builder{}
	bundles, err := os.ReadDir(cfg.UploadDir)

	if err != nil {
		slog.Error("Failed to list upload directory", "error", err)
		return 1
	}

	if len(bundles) == 0 {
		slog.Info("No Bundles found")
		return 0
	}

	buf.WriteString("Starting Deploy\nFound Bundles:")

	for _, b := range bundles {
		buf.WriteRune('\n')
		buf.WriteString(b.Name())
	}

	err = client.SendEmail(config.DeployStartEvent, buf.String())

	if err != nil {
		slog.Warn("Failed to send deploy start email", "error", err)
	}

	slog.Info("Unpacking bundles")
	lastRelease, err := unpack.UnpackBundles(
		cfg.UploadDir,
		cfg.ReleaseDir,
		cfg.UnpackScript,
		bundles,
		cfg.VersionConn,
	)

	if err != nil {
		slog.Error("Failed to unpack bundles", "error", err)
		err = client.SendEmail(config.DeployFinishEvent, "Failed to unpack deploy bundles")

		if err != nil {
			slog.Warn("Failed to send deploy error email", "error", err)
		}

		return 1
	}

	slog.Info("Setting up release", "final release", lastRelease)
	err = release.SetupRelease(cfg.ReleaseDir, lastRelease, cfg.ReloadScript, cfg.HealthScript)

	if err != nil {
		if errors.Is(err, release.RestartErr) {
			slog.Error("Failed restart", "error", err)
			err = client.SendEmail(config.DeployFinishEvent, "Restart failed")

			if err != nil {
				slog.Warn("Failed to send restart failure email", "error", err)
			}

			rollbackRet := rollback(cfg, client, 1)

			if rollbackRet == 1 {
				slog.Error("Automatic rollback Failed")
			}

			return 1
		}

		if errors.Is(err, release.HealthCheckErr) {
			slog.Error("Failed health check", "error", err)
			err = client.SendEmail(config.DeployFinishEvent, "Health check failed")

			if err != nil {
				slog.Warn("Failed to send health check failure email", "error", err)
			}

			rollbackRet := rollback(cfg, client, 1)

			if rollbackRet == 1 {
				slog.Error("Automatic rollback Failed")
			}

			return 1
		}

		slog.Error("Failed to set up new release", "error", err)
		err = client.SendEmail(config.DeployFinishEvent, "Failed to restart server")

		if err != nil {
			slog.Warn("Failed to send deploy restart email", "error", err)
		}

		return 1
	}

	err = client.SendEmail(config.DeployFinishEvent, "Deployed new server")

	return 0
}
