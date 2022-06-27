package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"reflect"
	"strings"
)

func configureLogger(version string, commit string) *Logger {
	return &Logger{
		tags: map[string]interface{}{
			"provider": "jmatsu/slack",
			"version":  version,
			"commit":   commit,
		},
		delegation: nil,
	}
}

type Logger struct {
	tags       map[string]interface{}
	delegation *Logger
}

func (logger *Logger) lift(ctx context.Context, level string, message string, tags ...map[string]interface{}) {
	if reflect.ValueOf(logger).IsNil() {
		switch strings.ToLower(level) {
		case "trace":
			tflog.Trace(ctx, message, tags...)
		case "debug":
			tflog.Debug(ctx, message, tags...)
		case "info":
			tflog.Info(ctx, message, tags...)
		case "warn":
			tflog.Warn(ctx, message, tags...)
		case "error":
			tflog.Error(ctx, message, tags...)
		}
		return
	}

	logger.delegation.lift(ctx, level, message, append(tags, logger.tags)...)
}

func (logger *Logger) withTags(tags map[string]interface{}) *Logger {
	return &Logger{
		tags:       tags,
		delegation: logger,
	}
}

func (logger *Logger) trace(ctx context.Context, format string, v ...interface{}) {
	logger.lift(ctx, "trace", fmt.Sprintf(format, v...))
}

func (logger *Logger) debug(ctx context.Context, format string, v ...interface{}) {
	logger.lift(ctx, "debug", fmt.Sprintf(format, v...))
}

func (logger *Logger) info(ctx context.Context, format string, v ...interface{}) {
	logger.lift(ctx, "info", fmt.Sprintf(format, v...))
}

func (logger *Logger) warning(ctx context.Context, format string, v ...interface{}) {
	logger.lift(ctx, "warn", fmt.Sprintf(format, v...))
}

func (logger *Logger) error(ctx context.Context, format string, v ...interface{}) {
	logger.lift(ctx, "error", fmt.Sprintf(format, v...))
}
