package models

import (
	"errors"
	"fmt"
)

// Common exit codes (keep small + consistent).
const (
	ExitOK       = 0
	ExitUsage    = 2 // invalid flags / bad CLI usage
	ExitRuntime  = 1 // generic runtime failure
	ExitConfig   = 3 // invalid config / validation failure
	ExitIO       = 4 // filesystem/IO issues
	ExitAuth     = 5 // auth failures
	ExitExternal = 6 // external dependency/service failure (e.g., Tor)
)

// CLIError is a user-facing error with optional hint + wrapped cause.
// Message/Hint are intended to be printed to the terminal.
type CLIError struct {
	Code     string // stable identifier for matching/logging (e.g. "CFG_INVALID")
	Message  string // user-facing message
	Hint     string // optional "try this"
	ExitCode int    // process exit code

	Cause error // underlying error (optional)
}

func (e *CLIError) Error() string {
	// Keep Error() concise and user-friendly.
	// Detailed internal cause can be logged separately.
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *CLIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *CLIError) WithHint(h string) *CLIError {
	if e == nil {
		return nil
	}
	e.Hint = h
	return e
}

func (e *CLIError) WithCause(err error) *CLIError {
	if e == nil {
		return nil
	}
	e.Cause = err
	return e
}

func NewCLIError(code string, exitCode int, msg string) *CLIError {
	return &CLIError{
		Code:    code,
		Message: msg,
		ExitCode: func() int {
			if exitCode == 0 {
				return ExitRuntime
			}
			return exitCode
		}(),
	}
}

// Wrap creates a CLIError while preserving an underlying cause.
func Wrap(code string, exitCode int, msg string, cause error) *CLIError {
	return NewCLIError(code, exitCode, msg).WithCause(cause)
}

// IsCode checks whether err (or any wrapped error) is a CLIError with the given code.
func IsCode(err error, code string) bool {
	var ce *CLIError
	if errors.As(err, &ce) {
		return ce.Code == code
	}
	return false
}

// FormatForUser builds the terminal output string for a CLIError.
// Use this for printing; keep logging separate.
func FormatForUser(err error) (text string, exitCode int) {
	if err == nil {
		return "", ExitOK
	}

	var ce *CLIError
	if errors.As(err, &ce) {
		exit := ce.ExitCode
		if exit == 0 {
			exit = ExitRuntime
		}

		if ce.Hint != "" {
			return fmt.Sprintf("%s\nhint: %s", ce.Message, ce.Hint), exit
		}
		return ce.Message, exit
	}

	// Non-CLIError fallback
	return err.Error(), ExitRuntime
}
