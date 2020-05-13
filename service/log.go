package service

import (
	"fmt"
	"github.com/cyolo-core/cmd/dex/repo"
)

// LogService allows creating entries in the log.
type LogService interface {
	// Logf logs a log entry with format string.
	Logf(username, remoteAddr, mapping, format string, a ...interface{})
}

// verify that logServiceImpl implements LogService.
var _ LogService = logServiceImpl{}

// logServiceImpl is the log service implementation
type logServiceImpl struct {
	repo.LogEntryRepository
}

// Logf logs a log entry.
func (svc logServiceImpl) Logf(username, remoteAddr, mapping, text string, a ...interface{}) {
	svc.Create(username, remoteAddr, mapping, fmt.Sprintf(text, a...))
}

// NewLogService returns a new LogService
func NewLogService(log repo.LogEntryRepository) LogService {
	return &logServiceImpl{log}
}
