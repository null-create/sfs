package logger

import "log/slog"

/*
Module for logging. Wrapper for slog package -- used
for the client, the server, and the monitor package
*/

type Logger struct {
	log *slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		log: &slog.Logger{},
	}
}
