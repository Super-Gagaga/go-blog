package slogx_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/yourname/go-blog/internal/pkg/slogx"
	"github.com/yourname/go-blog/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefault_ExplicitLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	l := slogx.NewDefault()
	assert.True(t, l.Enabled(context.Background(), slog.LevelDebug))
}

func TestNewDefault_TestProcessQuietByDefault(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	l := slogx.NewDefault()
	// 测试进程未显式配置:warn 生效、info 静默
	assert.True(t, l.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, l.Enabled(context.Background(), slog.LevelInfo))
}

func TestSetLogger_InjectAndReset(t *testing.T) {
	custom := slog.New(slog.NewTextHandler(&nullWriter{}, nil))
	slogx.SetLogger(custom)
	t.Cleanup(func() { slogx.SetLogger(nil) })
	assert.Same(t, custom, slogx.L())

	slogx.SetLogger(nil)
	assert.NotSame(t, custom, slogx.L())
}

type nullWriter struct{}

func (nullWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestRequestID_ContextRoundtrip(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, slogx.GetRequestID(ctx))

	ctx = slogx.WithRequestID(ctx, "req-42")
	assert.Equal(t, "req-42", slogx.GetRequestID(ctx))
}

func TestFromContext_AttachesRequestID(t *testing.T) {
	h := testutil.InstallCaptureLogger(t)
	ctx := slogx.WithRequestID(context.Background(), "req-42")

	slogx.FromContext(ctx).Warn("some.event", "key", "user:1")

	rec, ok := h.Find("some.event")
	require.True(t, ok, "event should be logged")
	v, ok := testutil.Attr(rec, "request_id")
	require.True(t, ok, "request_id attr should be present")
	assert.Equal(t, "req-42", v.String())
}

func TestFromContext_NoRequestID(t *testing.T) {
	h := testutil.InstallCaptureLogger(t)

	slogx.FromContext(context.Background()).Warn("plain.event")

	rec, ok := h.Find("plain.event")
	require.True(t, ok)
	_, hasAttr := testutil.Attr(rec, "request_id")
	assert.False(t, hasAttr, "no request_id attr when context has none")
}
