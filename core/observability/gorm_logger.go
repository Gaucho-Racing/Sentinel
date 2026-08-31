package observability

import (
	"context"
	"errors"
	"strings"
	"time"

	appLogger "github.com/gaucho-racing/sentinel/core/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseLogger struct {
	level         logger.LogLevel
	slowThreshold time.Duration
}

func NewDatabaseLogger(slowThreshold time.Duration) logger.Interface {
	return &DatabaseLogger{level: logger.Warn, slowThreshold: slowThreshold}
}

func (l *DatabaseLogger) LogMode(level logger.LogLevel) logger.Interface {
	copy := *l
	copy.level = level
	return &copy
}

func (l *DatabaseLogger) Info(_ context.Context, message string, args ...interface{}) {
	if l.level >= logger.Info {
		appLogger.SugarLogger.Infof(message, args...)
	}
}

func (l *DatabaseLogger) Warn(_ context.Context, message string, args ...interface{}) {
	if l.level >= logger.Warn {
		appLogger.SugarLogger.Warnf(message, args...)
	}
}

func (l *DatabaseLogger) Error(_ context.Context, message string, args ...interface{}) {
	if l.level >= logger.Error {
		appLogger.SugarLogger.Errorf(message, args...)
	}
}

func (l *DatabaseLogger) Trace(_ context.Context, started time.Time, query func() (string, int64), err error) {
	elapsed := time.Since(started)
	sql, rows := query()
	operation := databaseOperation(sql)
	slow := l.slowThreshold > 0 && elapsed >= l.slowThreshold
	ObserveDatabaseQuery(operation, err != nil, slow, elapsed)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= logger.Error {
		appLogger.SugarLogger.Errorf("Database query failed (operation=%s duration=%s rows=%d): %v", operation, elapsed, rows, err)
		return
	}
	if slow && l.level >= logger.Warn {
		appLogger.SugarLogger.Warnf("Slow database query (operation=%s duration=%s rows=%d)", operation, elapsed, rows)
	}
}

func databaseOperation(sql string) string {
	fields := strings.Fields(sql)
	if len(fields) == 0 {
		return "OTHER"
	}
	operation := strings.ToUpper(fields[0])
	switch operation {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "WITH", "CREATE", "ALTER", "DROP":
		return operation
	default:
		return "OTHER"
	}
}
