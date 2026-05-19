package slogx_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/yourname/go-blog/internal/pkg/slogx"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapHandler_ForwardsRecords(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	handler := slogx.NewZapHandler(zap.New(core))

	logger := slog.New(handler)
	logger.Warn("cache.invalidate_failed", "key", "user:1", "retries", 3)
	logger.Error("db.error", "err", assert.AnError)

	entries := logs.TakeAll()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	assert.Equal(t, "cache.invalidate_failed", entries[0].Message)
	assert.Equal(t, zap.WarnLevel, entries[0].Level)
	if field := entries[0].ContextMap()["key"]; field != "user:1" {
		t.Errorf("expected key=user:1, got %v", field)
	}
	if field := entries[0].ContextMap()["retries"]; field != int64(3) {
		t.Errorf("expected retries=3, got %v", field)
	}
	assert.Equal(t, zap.ErrorLevel, entries[1].Level)
}

func TestZapHandler_RespectsZapLevel(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel) // debug/info 被内核丢弃
	handler := slogx.NewZapHandler(zap.New(core))

	logger := slog.New(handler)
	logger.Info("dropped")
	logger.Warn("kept")

	entries := logs.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	assert.Equal(t, "kept", entries[0].Message)
}

func TestZapHandler_Enabled(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	handler := slogx.NewZapHandler(zap.New(core))
	assert.False(t, handler.Enabled(context.Background(), slog.LevelDebug))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))
}
