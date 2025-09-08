package logger

import (
	"time"
	"sync"
	"github.com/gin-gonic/gin"
)

// StandardLogger 标准日志记录器实现
type StandardLogger struct {
	mu        sync.RWMutex
	formatter LogFormatter
	writer    LogWriter
	filter    LogFilter
	enabled   bool
}

// NewStandardLogger 创建新的标准日志记录器
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		formatter: NewTextFormatter(),
		writer:    NewConsoleWriter(),
		filter:    NewCompositeFilter(FilterModeAnd), // 默认使用组合过滤器
		enabled:   true,
	}
}

// Log 记录日志条目
func (l *StandardLogger) Log(entry *LogEntry) error {
	if !l.enabled {
		return nil
	}
	
	l.mu.RLock()
	writer := l.writer
	l.mu.RUnlock()
	
	return writer.Write(entry)
}

// LogAccess 记录访问日志
func (l *StandardLogger) LogAccess(c *gin.Context, statusCode int, responseTime time.Duration) {
	if !l.enabled {
		return
	}
	
	l.mu.RLock()
	filter := l.filter
	l.mu.RUnlock()
	
	// 检查过滤器
	if filter != nil && !filter.ShouldLog(c, TypeAccess) {
		return
	}
	
	entry := &LogEntry{
		Timestamp:    time.Now(),
		Level:        LevelInfo,
		Type:         TypeAccess,
		IP:           c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		StatusCode:   statusCode,
		ResponseTime: responseTime.Milliseconds(),
		Referrer:     c.GetHeader("Referer"),
	}
	
	// 尝试获取用户名
	if username, exists := c.Get("username"); exists {
		if str, ok := username.(string); ok {
			entry.Username = str
		}
	}
	
	l.Log(entry)
}

// LogMedia 记录媒体文件访问日志
func (l *StandardLogger) LogMedia(c *gin.Context, username string, fileSize int64) {
	if !l.enabled {
		return
	}
	
	l.mu.RLock()
	filter := l.filter
	l.mu.RUnlock()
	
	// 检查过滤器
	if filter != nil && !filter.ShouldLog(c, TypeMedia) {
		return
	}
	
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LevelInfo,
		Type:      TypeMedia,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Username:  username,
		FileSize:  fileSize,
		Referrer:  c.GetHeader("Referer"),
	}
	
	l.Log(entry)
}

// LogError 记录错误日志
func (l *StandardLogger) LogError(c *gin.Context, err error, message string) {
	if !l.enabled {
		return
	}
	
	l.mu.RLock()
	filter := l.filter
	l.mu.RUnlock()
	
	// 检查过滤器
	if filter != nil && !filter.ShouldLog(c, TypeError) {
		return
	}
	
	errorMsg := message
	if err != nil {
		if errorMsg != "" {
			errorMsg += ": " + err.Error()
		} else {
			errorMsg = err.Error()
		}
	}
	
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LevelError,
		Type:      TypeError,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Message:   errorMsg,
	}
	
	// 尝试获取用户名
	if username, exists := c.Get("username"); exists {
		if str, ok := username.(string); ok {
			entry.Username = str
		}
	}
	
	l.Log(entry)
}

// LogSystem 记录系统日志
func (l *StandardLogger) LogSystem(level LogLevel, message string, extra map[string]interface{}) {
	if !l.enabled {
		return
	}
	
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Type:      TypeSystem,
		Message:   message,
		Extra:     extra,
	}
	
	l.Log(entry)
}

// SetFilter 设置过滤器
func (l *StandardLogger) SetFilter(filter LogFilter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filter = filter
}

// SetWriter 设置写入器
func (l *StandardLogger) SetWriter(writer LogWriter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// 关闭旧的写入器
	if l.writer != nil {
		l.writer.Close()
	}
	
	l.writer = writer
}

// SetFormatter 设置格式化器
func (l *StandardLogger) SetFormatter(formatter LogFormatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = formatter
}

// SetEnabled 设置是否启用日志
func (l *StandardLogger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// Close 关闭日志记录器
func (l *StandardLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}

// GetFormatter 获取当前格式化器
func (l *StandardLogger) GetFormatter() LogFormatter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.formatter
}

// GetWriter 获取当前写入器
func (l *StandardLogger) GetWriter() LogWriter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.writer
}

// GetFilter 获取当前过滤器
func (l *StandardLogger) GetFilter() LogFilter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.filter
}

// IsEnabled 检查是否启用
func (l *StandardLogger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}
