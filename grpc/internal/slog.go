// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Slogger is a grpclog.LoggerV2 implementation with slog.Logger.
// It's used to replace the default grpclog.LoggerV2:
//
//	grpclog.SetLoggerV2(NewSlogger(handler))
//
// To create a new Slogger, use [NewSlogger].
type Slogger struct {
	handler   slog.Handler
	severity  slog.Level
	verbosity int
}

// NewSlogger creates a new Slogger with the given slog.Handler.
func NewSlogger(handler slog.Handler) Slogger {
	var severity slog.Level
	switch strings.ToLower(os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL")) {
	case "warn", "warning":
		severity = slog.LevelWarn
	case "info":
		severity = slog.LevelInfo
	default:
		severity = slog.LevelError
	}
	verbosity, _ := strconv.Atoi(os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL"))

	return Slogger{
		handler:   handler,
		severity:  severity,
		verbosity: verbosity,
	}
}

func (g Slogger) Info(args ...any) {
	g.log(0, slog.LevelInfo, fmt.Sprint(args...))
}

func (g Slogger) Infoln(args ...any) {
	g.log(0, slog.LevelInfo, fmt.Sprint(args...))
}

func (g Slogger) Infof(format string, args ...any) {
	g.log(0, slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (g Slogger) InfoDepth(depth int, args ...any) {
	g.log(depth, slog.LevelInfo, fmt.Sprint(args...))
}

func (g Slogger) Warning(args ...any) {
	g.log(0, slog.LevelWarn, fmt.Sprint(args...))
}

func (g Slogger) Warningln(args ...any) {
	g.log(0, slog.LevelWarn, fmt.Sprint(args...))
}

func (g Slogger) Warningf(format string, args ...any) {
	g.log(0, slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (g Slogger) WarningDepth(depth int, args ...any) {
	g.log(depth, slog.LevelWarn, fmt.Sprint(args...))
}

func (g Slogger) Error(args ...any) {
	g.log(0, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) Errorln(args ...any) {
	g.log(0, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) Errorf(format string, args ...any) {
	g.log(0, slog.LevelError, fmt.Sprintf(format, args...))
}

func (g Slogger) ErrorDepth(depth int, args ...any) {
	g.log(depth, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) Fatal(args ...any) {
	g.log(0, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) Fatalln(args ...any) {
	g.log(0, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) Fatalf(format string, args ...any) {
	g.log(0, slog.LevelError, fmt.Sprintf(format, args...))
}

func (g Slogger) FatalDepth(depth int, args ...any) {
	g.log(depth, slog.LevelError, fmt.Sprint(args...))
}

func (g Slogger) V(l int) bool {
	return l <= g.verbosity
}

func (g Slogger) log(depth int, level slog.Level, message string) {
	if g.severity > level {
		return
	}

	handler := g.handler
	if g.handler == nil {
		handler = slog.Default().Handler()
	}

	ctx := context.Background()
	if !handler.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// Skip runtime.Callers, this method, log methods like Info in this package and grpclog.
	runtime.Callers(depth+4, pcs[:]) //nolint:gomnd
	// Ignore error: It's fine to lose log.
	_ = handler.Handle(ctx, slog.NewRecord(time.Now(), level, message, pcs[0]))
}
