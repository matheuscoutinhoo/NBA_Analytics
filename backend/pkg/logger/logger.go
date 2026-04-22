package logger

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
)

type Logger struct {
	output io.Writer
	level  Level
	mu     sync.Mutex
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     Level                  `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var levelPriority = map[Level]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
	LevelFatal: 4,
}

func New(output io.Writer, level Level) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		output: output,
		level:  level,
	}
}

func Default() *Logger {
	return New(os.Stdout, LevelInfo)
}

func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	if levelPriority[level] < levelPriority[l.level] {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	data = append(data, '\n')
	l.output.Write(data)

	if level == LevelFatal {
		os.Exit(1)
	}
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(LevelDebug, message, mergeFields(fields))
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(LevelInfo, message, mergeFields(fields))
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(LevelWarn, message, mergeFields(fields))
}

func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(LevelError, message, mergeFields(fields))
}

func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	l.log(LevelFatal, message, mergeFields(fields))
}

func mergeFields(fields []map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}
