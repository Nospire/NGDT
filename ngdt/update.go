package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func RunUpdate() error {
	cmd := exec.Command("steamos-update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 7 {
			return nil // exit 7 = no updates available, treat as success
		}
		return fmt.Errorf("steamos-update: %w", err)
	}
	return nil
}
