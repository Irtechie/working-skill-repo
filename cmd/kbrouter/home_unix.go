//go:build !windows

package main

import (
	"fmt"
	"os/user"
)

func operatingSystemUserHome() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("resolve current user home: %w", err)
	}
	if current.HomeDir == "" {
		return "", fmt.Errorf("resolve current user home: empty path")
	}
	return current.HomeDir, nil
}
