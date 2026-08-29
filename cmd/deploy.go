package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/Dannie226/cd_watcher/internal/email"
	"github.com/Dannie226/cd_watcher/internal/release"
	"github.com/Dannie226/cd_watcher/internal/unpack"
	"github.com/jackc/pgx/v5"
)

func deploy() int {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()

	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return 1
	}

	credDir, ok := os.LookupEnv("CREDENTIALS_DIR")

	if !ok {
		slog.Error("No credentials directory environment variable")

		// For now, not exiting, just going to set it to a local value for testing
		// os.Exit(1)
		credDir = "./creds"
	}

	dbUrl, err := os.ReadFile(fmt.Sprintf("%s/pg_url", credDir))

	if err != nil {
		slog.Error("Failed to read postgres database url", "error", err)
		return 1
	}

	pgUrl := strings.TrimSpace(string(dbUrl))

	emailConn, err := pgx.Connect(context.Background(), pgUrl)

	if err != nil {
		slog.Error("Failed to create email database connection", "error", err)
		return 1
	}
	defer emailConn.Close(context.Background())

	var client *email.EmailClient

	if cfg.EmailConfig != nil {
		client, err = email.NewClient(cfg.EmailConfig, credDir, emailConn)

		if err != nil {
			slog.Error("Failed to create email client", "error", err)
			return 1
		}
	}

	uploadConn, err := pgx.Connect(context.Background(), pgUrl)

	if err != nil {
		slog.Error("Failed to create upload database connection", "error", err)
		return 1
	}

	defer uploadConn.Close(context.Background())

	client.SendEmail(config.DeployStartEvent, "Starting deploy")

	lastRelease, err := unpack.UnpackBundles(cfg.UploadDir, cfg.ReleaseDir, cfg.UnpackScript, uploadConn)

	if err != nil {
		if errors.Is(err, unpack.ErrNoBundles) {
			slog.Info(err.Error())
			err := client.SendEmail(config.DeployFinishEvent, "No bundles were found for deploying")

			if err != nil {
				slog.Error("Failed to send deploy finish email", "error", err)
				return 1
			}

			return 0
		}

		slog.Error("Failed to unpack bundles", "error", err)
		client.SendEmail(config.DeployFinishEvent, "Failed to unpack deploy bundles")
		return 1
	}

	err = release.SetupRelease(cfg.ReleaseDir, lastRelease, cfg.ReloadScript, cfg.HealthScript)

	if err != nil {
		slog.Error("Failed to set up new release", "error", err)
		client.SendEmail(config.DeployFinishEvent, "Failed to restart server")
		return 1
	}

	return 0
}
