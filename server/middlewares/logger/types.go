package logger

import (
	"time"
	"github.com/gin-gonic/gin"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// LogType 日志类型
type LogType string

const (
	TypeAccess LogType = "access"  // 访问日志
	TypeMedia  LogType = "media"   // 媒体文件访问日志
	TypeError  LogType = "error"   // 错误日志
	TypeSystem LogType = "system"  // 系统日志
)

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       LogLevel  `json:"level"`
	Type        LogType   `json:"type"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code,omitempty"`
	ResponseTime int64    `json:"response_time,omitempty"` // 毫秒
	Username    string    `json:"username,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	Referrer    string    `json:"referrer,omitempty"`
	Message     string    `json:"message,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// LogFormatter 日志格式化器接口
type LogFormatter interface {
	Format(entry *LogEntry) string
}

// LogWriter 日志写入器接口
type LogWriter interface {
	Write(entry *LogEntry) error
	Close() error
}

// LogFilter 日志过滤器接口
type LogFilter interface {
	ShouldLog(c *gin.Context, logType LogType) bool
}

// Logger 统一日志记录器接口
type Logger interface {
	Log(entry *LogEntry) error
	LogAccess(c *gin.Context, statusCode int, responseTime time.Duration)
	LogMedia(c *gin.Context, username string, fileSize int64)
	LogError(c *gin.Context, err error, message string)
	LogSystem(level LogLevel, message string, extra map[string]interface{})
	SetFilter(filter LogFilter)
	SetWriter(writer LogWriter)
	SetFormatter(formatter LogFormatter)
}
