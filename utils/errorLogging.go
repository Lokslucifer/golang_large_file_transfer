package utils
import (
	"context"
	"log/slog"
	"runtime/debug"
)
func LogErrorWithStack(ctx context.Context, msg string, err error, attrs ...interface{}) {
	stackTrace := string(debug.Stack()) // Get the stack trace

	// Create a slice to hold all attributes, starting with the error
	allAttrs := []interface{}{
		slog.String("error", err.Error()),
		slog.String("stack_trace", stackTrace), // Add the stack trace
	}

	// Add any user-provided attributes
	allAttrs = append(allAttrs, attrs...)

	// Use slog.ErrorContext with the combined attributes.
	slog.ErrorContext(ctx, msg, allAttrs...)
}