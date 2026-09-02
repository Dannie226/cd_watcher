package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"strconv"

	"github.com/Dannie226/cd_watcher/internal/command"
	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/Dannie226/cd_watcher/internal/email"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	if os.Args[1] != "deploy" && os.Args[1] != "rollback" {
		usage()
		os.Exit(1)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()

	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.StopScript != "" {
		if err := command.RunCommand(
			"/usr/bin/bash",
			"",
			"Stop",
			cfg.StopScript,
		); err != nil {
			slog.Error("Failed to stop service", "error", err)
			os.Exit(1)
		}
	}

	var client *email.EmailClient

	if cfg.EmailConfig != nil {
		client = email.NewClient(cfg.EmailConfig, cfg.EmailLogin, cfg.EmailConn)
	}

	switch os.Args[1] {
	case "deploy":
		ret := deploy(cfg, client)

		cfg.EmailConn.Close(context.Background())
		cfg.VersionConn.Close(context.Background())

		os.Exit(ret)

	case "rollback":
		if len(os.Args) < 3 {
			fmt.Println("Rollback command must have a positive numeric argument")
			os.Exit(1)
		}

		num, err := strconv.Atoi(os.Args[2])

		if err != nil || num <= 0 {
			fmt.Println("Rollback command argument must be a positive number")
			os.Exit(1)
		}

		ret := rollback(cfg, client, num)

		cfg.EmailConn.Close(context.Background())
		cfg.VersionConn.Close(context.Background())

		os.Exit(ret)
	}
}
