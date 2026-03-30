package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mcp-notify/internal/mcp"
	"mcp-notify/internal/player"
	"mcp-notify/internal/validation"
)

func main() {
	logger := log.New(os.Stderr, "mcp-notify: ", log.LstdFlags|log.Lmsgprefix)

	workingDir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("failed to determine working directory: %v", err)
	}
	baseDir, err := resolveBaseDir(workingDir)
	if err != nil {
		logger.Fatalf("failed to determine base directory: %v", err)
	}

	config, err := parseConfig(os.Args[1:])
	if err != nil {
		logger.Fatalf("failed to parse startup configuration: %v", err)
	}

	if config.playOncePath != "" {
		if err := player.New().Play(context.Background(), config.playOncePath, true); err != nil {
			logger.Fatalf("failed to play detached sound: %v", err)
		}
		return
	}
	validator := validation.NewSoundValidator(baseDir)
	if config.playOnceSound != "" {
		resolved, err := validator.ValidateOneShotPath(config.playOnceSound)
		if err != nil {
			logger.Fatalf("failed to validate play-once sound: %v", err)
		}
		if err := player.New().Play(context.Background(), resolved.ResolvedPath, config.wait); err != nil {
			logger.Fatalf("failed to play one-shot sound: %v", err)
		}
		return
	}

	server := mcp.NewServer(config.serverName, "0.1.0", logger)
	server.SetInitializeCheck(func() *mcp.ResponseError {
		if config.soundPath == "" {
			return nil
		}
		if _, err := validator.ValidateConfiguredPath(config.soundPath); err != nil {
			return &mcp.ResponseError{
				Code:    -32602,
				Message: "invalid startup sound configuration",
				Data: map[string]any{
					"error":   err.Message,
					"details": err.Details,
				},
			}
		}
		return nil
	})
	server.RegisterTool(mcp.NewPlayNotificationSoundTool(
		validator,
		player.New(),
		config.soundPath,
		config.wait,
		config.toolPrefix,
	))

	if err := server.Serve(context.Background(), os.Stdin, os.Stdout); err != nil {
		logger.Fatalf("server stopped with error: %v", err)
	}
}

type config struct {
	soundPath     string
	wait          bool
	playOnceSound string
	playOncePath  string
	serverName    string
	toolPrefix    string
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("mcp-notify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var cfg config
	fs.StringVar(&cfg.soundPath, "sound", "", "file name or relative path under the sounds directory")
	fs.BoolVar(&cfg.wait, "wait", true, "wait until playback completes before returning from the tool")
	fs.StringVar(&cfg.playOnceSound, "play-once", "", "play one sound file under the sounds directory and exit")
	fs.StringVar(&cfg.playOncePath, "play-once-path", "", "internal: absolute sound path to play once and exit")
	fs.StringVar(&cfg.serverName, "server-name", "mcp-notify", "name returned from initialize.serverInfo.name")
	fs.StringVar(&cfg.toolPrefix, "tool-prefix", "", "literal prefix added to the exposed MCP tool name")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if cfg.playOnceSound != "" && cfg.soundPath != "" {
		return config{}, fmt.Errorf("--play-once cannot be combined with --sound")
	}
	if cfg.playOnceSound != "" && cfg.playOncePath != "" {
		return config{}, fmt.Errorf("--play-once cannot be combined with --play-once-path")
	}

	return cfg, nil
}

func resolveBaseDir(workingDir string) (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return resolveBaseDirFromExecutable(workingDir, executablePath)
}

func resolveBaseDirFromExecutable(workingDir, executablePath string) (string, error) {
	executableDir := filepath.Dir(executablePath)
	for _, candidate := range []string{
		executableDir,
		filepath.Dir(executableDir),
	} {
		if info, err := os.Stat(filepath.Join(candidate, "sounds")); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return workingDir, nil
}
