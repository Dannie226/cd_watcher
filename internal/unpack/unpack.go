package unpack

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

var ErrNoBundles error = fmt.Errorf("No bundles found for deploy")

func UnpackBundles(uploadDir, releaseDir, unpackScript string, uploadConn *pgx.Conn) (string, error) {
	entries, err := os.ReadDir(uploadDir)

	if err != nil {
		return "", fmt.Errorf("Failed to read upload directory: %w", err)
	}

	if len(entries) == 0 {
		return "", ErrNoBundles
	}

	var relName string

	for _, e := range entries {
		randBytes := [16]byte{0}
		rand.Read(randBytes[:])

		relName = base64.RawURLEncoding.EncodeToString(randBytes[:])

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

		err = os.Remove(entryName)

		if err != nil {
			return "", fmt.Errorf("Failed to remove entry")
		}
	}

	return relName, nil
}
