package command

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
)

func RunCommand(command string, workingDir string, logName string, args ...string) error {
	cmd := exec.Command(command, args...)

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	errPipe, err := cmd.StderrPipe()

	if err != nil {
		return fmt.Errorf("Failed to get command stderror pipe: %w", err)
	}

	defer errPipe.Close()

	outPipe, err := cmd.StdoutPipe()

	if err != nil {
		return fmt.Errorf("Failed to get command stdout pipe: %w", err)
	}

	defer outPipe.Close()

	errScanner := bufio.NewScanner(errPipe)
	outScanner := bufio.NewScanner(outPipe)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start running command: %w", err)
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		for errScanner.Scan() {
			t := errScanner.Text()

			slog.Warn("Command Stderr", "output", t, "command", command, "log name", logName)
		}

		if err := errScanner.Err(); err != nil {
			slog.Error("Failed to read command stderr", "error", err)
			io.Copy(io.Discard, errPipe)
		}
	})

	wg.Go(func() {
		for outScanner.Scan() {
			t := outScanner.Text()

			slog.Info("Command Stdout", "output", t, "command", command, "log name", logName)
		}

		if err := outScanner.Err(); err != nil {
			slog.Error("Failed to read command stdout", "error", err)
			io.Copy(io.Discard, outPipe)
		}
	})

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Failed to run command: %w", err)
	}

	return nil
}
