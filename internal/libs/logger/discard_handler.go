package logger

import (
	"context"
	"log/slog"
)

// TODO: Remove this file and the handler when the default [slog.DiscardHandler] will be introduced in
//   Go version 1.24. See https://go-review.googlesource.com/c/go/+/626486.var DiscardHandler slog.Handler = discardHandler{}

var DiscardHandler slog.Handler = discardHandler{}

type discardHandler struct{}

func (dh discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (dh discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (dh discardHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return dh }
func (dh discardHandler) WithGroup(name string) slog.Handler        { return dh }
