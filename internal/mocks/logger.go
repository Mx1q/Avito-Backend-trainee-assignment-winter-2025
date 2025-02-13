package mocks

import "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"

type mockLogger struct{}

func NewMockLogger() logger.ILogger {
	return &mockLogger{}
}

func (m *mockLogger) Infof(message string, args ...interface{}) {}

func (m *mockLogger) Warnf(message string, args ...interface{}) {}

func (m *mockLogger) Errorf(message string, args ...interface{}) {}

func (m *mockLogger) Fatalf(message string, args ...interface{}) {}

func (m *mockLogger) Debugf(message string, args ...interface{}) {}
