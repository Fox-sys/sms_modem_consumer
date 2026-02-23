package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"sender-modem/src/internal/application"
	"sender-modem/src/internal/infrastructure/api"
	"sender-modem/src/internal/infrastructure/huawei"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := application.LoadConfig()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if strings.EqualFold(cfg.LogLevel, "debug") {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	reader := huawei.NewAdapter(cfg.ModemBaseURL, cfg.ModemUsername, cfg.ModemPassword)
	forwarder := api.NewClient(cfg.APIBaseURL, cfg.APIKey)

	uc := &application.PollAndForward{
		Reader:    reader,
		Forwarder: forwarder,
		Interval:  time.Duration(cfg.PollIntervalSeconds) * time.Second,
	}

	if err := uc.Run(); err != nil {
		slog.Error("run failed", "err", err)
		os.Exit(1)
	}
}
