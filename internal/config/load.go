package config

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/jackc/pgx/v5"
)

type anonymousLogin struct {
	Trace string `json:"trace"`
}

type externalLogin struct {
	Identity string `json:"identity"`
}

type oauthLogin struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type plainLogin struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func parseLogin(login loginType, loginCreds []byte) (sasl.Client, error) {
	var c sasl.Client

	switch login {
	case Anonymous:
		a := anonymousLogin{}
		err := json.Unmarshal(loginCreds, &a)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal anonymous credentials: %w", err)
		}

		c = sasl.NewAnonymousClient(a.Trace)

	case External:
		e := externalLogin{}
		err := json.Unmarshal(loginCreds, &e)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal external credentials: %w", err)
		}

		c = sasl.NewExternalClient(e.Identity)

	case OAuth:
		e := oauthLogin{}
		err := json.Unmarshal(loginCreds, &e)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal OAuth credentials: %w", err)
		}

		opts := sasl.OAuthBearerOptions{
			Username: e.Username,
			Token:    e.Token,
			Host:     e.Host,
			Port:     e.Port,
		}

		c = sasl.NewOAuthBearerClient(&opts)

	case Plain:
		p := plainLogin{}
		err := json.Unmarshal(loginCreds, &p)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal plain credentials: %w", err)
		}

		c = sasl.NewPlainClient(p.Identity, p.Username, p.Password)
	}

	return c, nil
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ReleaseDir:   "releases",
		UploadDir:    "uploads",
		UnpackScript: "scripts/unpack.sh",
		ReloadScript: "scripts/reload.sh",
	}

	data, err := os.ReadFile("config.json")

	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %w", err)
	}

	err = json.Unmarshal(data, cfg)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	credDir, ok := os.LookupEnv("CREDENTIALS_DIR")

	if !ok {
		return nil, fmt.Errorf("No credentials directory environment variable")
	}

	var l sasl.Client

	if cfg.EmailConfig != nil {
		params, err := os.ReadFile(filepath.Join(credDir, "email_login"))

		if err != nil {
			return nil, fmt.Errorf("Failed to read email login parameters: %w", err)
		}

		c, err := parseLogin(cfg.EmailConfig.LoginType, params)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse email login parameters: %w", err)
		}

		l = c
	}

	cfg.EmailLogin = l

	url, err := os.ReadFile(filepath.Join(credDir, "pg_url"))

	if err != nil {
		return nil, fmt.Errorf("Failed to read database url: %w", err)
	}

	pgUrl := strings.TrimSpace(string(url))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	eConn, err := pgx.Connect(ctx, pgUrl)
	cancel()

	if err != nil {
		return nil, fmt.Errorf("Failed to create email database connection: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)

	vConn, err := pgx.Connect(ctx, pgUrl)
	cancel()

	if err != nil {
		eConn.Close(context.Background())

		return nil, fmt.Errorf("Failed to create version database connection: %w", err)
	}

	cfg.EmailConn = eConn
	cfg.VersionConn = vConn

	return cfg, nil
}
