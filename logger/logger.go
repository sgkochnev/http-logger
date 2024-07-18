package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type ctxKey string

const (
	pSlogFields ctxKey = "p_slog_fields"
	slogFields  ctxKey = "slog_fields"
)

type SlogHandler struct {
	next slog.Handler
}

func NewSlogHandler(next slog.Handler) *SlogHandler {
	return &SlogHandler{next: next}
}

func InitLogger(appName, version string, logLevel string) error {
	var level slog.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return fmt.Errorf("failed to parse log level (%s): %w", logLevel, err)
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: level,
	})

	slogHandler := NewSlogHandler(h.WithAttrs([]slog.Attr{
		slog.String("name", appName),
		slog.String("version", version),
	}))

	slog.SetDefault(slog.New(slogHandler))
	return nil
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		record.AddAttrs(attrs...)
	}
	if attrs, ok := ctx.Value(pSlogFields).(*[]slog.Attr); ok {
		record.AddAttrs(*attrs...)
	}
	return h.next.Handle(ctx, record)
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SlogHandler{h.next.WithAttrs(attrs)}
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return &SlogHandler{h.next.WithGroup(name)}
}

// CtxWith добавляет атрибуты в контекст возвращая новый контекст
func CtxWith(ctx context.Context, fields ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	slogAttrs, _ := ctx.Value(slogFields).([]slog.Attr)
	slogAttrs = append(slogAttrs, fields...)

	return context.WithValue(ctx, slogFields, slogAttrs)
}

// With добавляет атрибуты в контекст
func With(ctx context.Context, fields ...slog.Attr) {
	if !HasFieldInCtx(ctx) {
		return
	}

	slogAttrs, _ := ctx.Value(pSlogFields).(*[]slog.Attr)
	*slogAttrs = append(*slogAttrs, fields...)
}

func CtxInit(ctx context.Context) context.Context {
	if HasFieldInCtx(ctx) {
		return ctx
	}

	return context.WithValue(ctx, pSlogFields, &[]slog.Attr{})
}

func HasFieldInCtx(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(pSlogFields).(*[]slog.Attr)
	return ok && v != nil
}

func Group(ctx context.Context, group string, attrs ...slog.Attr) {
	if !HasFieldInCtx(ctx) {
		return
	}

	attr := slog.Group(group, attrs[0].Key, attrs[0].Value)
	With(ctx, attr)
}
