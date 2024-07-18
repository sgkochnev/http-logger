package logger

import (
	"context"
	"errors"
	"log/slog"
)

type ErrorWithContext struct {
	context.Context
	error
}

func Err(err error) error {
	if !errors.As(err, &ErrorWithContext{}) {
		return err
	}

	return err.(ErrorWithContext).error
}

func CtxFromErr(ctx context.Context, err error) context.Context {
	if !errors.As(err, &ErrorWithContext{}) {
		return ctx
	}

	fields, ok := err.(ErrorWithContext).Context.Value(slogFields).([]slog.Attr)
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, slogFields, fields)
}


func WrapError(ctx context.Context, err error) error {
	return ErrorWithContext{ctx, err}
}
