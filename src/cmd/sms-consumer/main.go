package main

import (
	"log/slog"
	"os"
	"time"

	"sender-modem/internal/application"
	"sender-modem/internal/infrastructure/api"
	"sender-modem/internal/infrastructure/huawei"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := application.LoadConfig()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

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
