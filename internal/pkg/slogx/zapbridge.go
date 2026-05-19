package slogx

import (
	"context"
	"log/slog"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapHandler 把 slog 记录桥接到 zap.Logger,使缓存/数据访问层
// (经 SetLogger 注入的 slog logger)的日志汇入应用统一的 zap 输出。
type ZapHandler struct {
	logger *zap.Logger
	groups []string
}

// NewZapHandler 返回包装 l 的 slog.Handler。
func NewZapHandler(l *zap.Logger) *ZapHandler {
	return &ZapHandler{logger: l}
}

func (h *ZapHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.Core().Enabled(zapLevel(level))
}

func (h *ZapHandler) Handle(_ context.Context, r slog.Record) error {
	ce := h.logger.Check(zapLevel(r.Level), r.Message)
	if ce == nil {
		return nil
	}
	fields := make([]zap.Field, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields = append(fields, h.field(a))
		return true
	})
	ce.Write(fields...)
	return nil
}

func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	clone := *h
	fields := make([]zap.Field, 0, len(attrs))
	for _, a := range attrs {
		fields = append(fields, h.field(a))
	}
	clone.logger = h.logger.With(fields...)
	return &clone
}

func (h *ZapHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.groups = append(append([]string{}, h.groups...), name)
	return &clone
}

func (h *ZapHandler) field(a slog.Attr) zap.Field {
	key := a.Key
	if len(h.groups) > 0 {
		key = strings.Join(h.groups, ".") + "." + a.Key
	}
	switch a.Value.Kind() {
	case slog.KindString:
		return zap.String(key, a.Value.String())
	case slog.KindInt64:
		return zap.Int64(key, a.Value.Int64())
	case slog.KindFloat64:
		return zap.Float64(key, a.Value.Float64())
	case slog.KindBool:
		return zap.Bool(key, a.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(key, a.Value.Duration())
	case slog.KindTime:
		return zap.Time(key, a.Value.Time())
	case slog.KindAny:
		if err, ok := a.Value.Any().(error); ok {
			return zap.NamedError(key, err)
		}
		return zap.Any(key, a.Value.Any())
	default:
		return zap.String(key, a.Value.String())
	}
}

func zapLevel(level slog.Level) zapcore.Level {
	switch {
	case level >= slog.LevelError:
		return zap.ErrorLevel
	case level >= slog.LevelWarn:
		return zap.WarnLevel
	case level >= slog.LevelInfo:
		return zap.InfoLevel
	default:
		return zap.DebugLevel
	}
}
