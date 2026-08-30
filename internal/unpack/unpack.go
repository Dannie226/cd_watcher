package unpack

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Dannie226/cd_watcher/internal/queries"
	"github.com/jackc/pgx/v5"
)

func UnpackBundles(uploadDir, releaseDir, unpackScript string, bundles []os.DirEntry, uploadConn *pgx.Conn) (string, error) {
	id, err := queries.GetNextVersionID(uploadConn)

	if err != nil {
		return "", fmt.Errorf("Failed to get next release id: %w", err)
	}

	var relName string

	for _, e := range bundles {
		randBytes := [24]byte{0}
		rand.Read(randBytes[:])

		relName = base64.RawURLEncoding.EncodeToString(randBytes[:])

		slog.Info("Unpacking bundle", "bundle name", e.Name(), "release folder", relName)

		relDir := filepath.Join(releaseDir, relName, "")
		entryName := filepath.Join(uploadDir, e.Name())

		relEntry, err := filepath.Rel(relDir, entryName)

		if err != nil {
			return "", fmt.Errorf("Failed to get relative path for bundle: %w", err)
		}

		relUnpack, err := filepath.Rel(relDir, unpackScript)

		if err != nil {
			return "", fmt.Errorf("Failed to get relative path for unpack script: %w", err)
		}

		absRelease, err := filepath.Abs(relDir)

		if err != nil {
			return "", fmt.Errorf("Failed to create absolute path for release directory: %w", err)
		}

		err = os.Mkdir(relDir, 0777)

		if err != nil {
			return "", fmt.Errorf("Failed to create release directory for \"%s\": %w", e.Name(), err)
		}

		cmd := exec.Command("/usr/bin/bash", relUnpack, relEntry)

		cmd.Dir = absRelease

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("Failed to run unpack script on \"%s\": %w", e.Name(), err)
		}

		queries.InsertNewVersion(uploadConn, queries.VersionInfo{
			ID:         id,
			FolderName: relName,
		})

		if err != nil {
			return "", fmt.Errorf("Failed to insert release into database: %w", err)
		}

		id++

		err = os.Remove(entryName)

		if err != nil {
			return "", fmt.Errorf("Failed to remove entry")
		}
	}

	return relName, nil
}
