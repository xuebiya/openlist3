package logger

import (
	"fmt"
	"sync"
)

// LoggerManager 日志管理器 - 管理多种类型的日志记录器
type LoggerManager struct {
	mu      sync.RWMutex
	loggers map[LogType]Logger
	config  *LogConfig
}

// LogConfig 日志配置
type LogConfig struct {
	Enabled         bool          `json:"enabled"`
	Level           LogLevel      `json:"level"`
	Format          string        `json:"format"`          // "text" 或 "json"
	Output          OutputConfig  `json:"output"`
	AccessLog       AccessConfig  `json:"access_log"`
	MediaLog        MediaConfig   `json:"media_log"`
	ErrorLog        ErrorConfig   `json:"error_log"`
	SystemLog       SystemConfig  `json:"system_log"`
}

type OutputConfig struct {
	Console ConsoleConfig `json:"console"`
	File    FileConfig    `json:"file"`
}

type ConsoleConfig struct {
	Enabled    bool `json:"enabled"`
	Colors     bool `json:"colors"`
}

type FileConfig struct {
	Enabled    bool   `json:"enabled"`
	Filename   string `json:"filename"`
	MaxSize    int64  `json:"max_size"`    // MB
	MaxFiles   int    `json:"max_files"`
	Async      bool   `json:"async"`
	BufferSize int    `json:"buffer_size"`
}

type AccessConfig struct {
	Enabled       bool     `json:"enabled"`
	ExcludePaths  []string `json:"exclude_paths"`
	IncludePaths  []string `json:"include_paths"`
	ExcludeStatus []int    `json:"exclude_status"`
	IncludeStatus []int    `json:"include_status"`
}

type MediaConfig struct {
	Enabled bool `json:"enabled"`
	// 媒体日志特定配置可以在这里添加
}

type ErrorConfig struct {
	Enabled bool `json:"enabled"`
}

type SystemConfig struct {
	Enabled bool `json:"enabled"`
}

// NewLoggerManager 创建新的日志管理器
func NewLoggerManager(config *LogConfig) *LoggerManager {
	if config == nil {
		config = DefaultLogConfig()
	}
	
	manager := &LoggerManager{
		loggers: make(map[LogType]Logger),
		config:  config,
	}
	
	manager.initializeLoggers()
	return manager
}

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Enabled: true,
		Level:   LevelInfo,
		Format:  "text",
		Output: OutputConfig{
			Console: ConsoleConfig{
				Enabled: true,
				Colors:  true,
			},
			File: FileConfig{
				Enabled:    true,
				Filename:   "./logs/openlist.log",
				MaxSize:    100, // 100MB
				MaxFiles:   10,
				Async:      true,
				BufferSize: 1000,
			},
		},
		AccessLog: AccessConfig{
			Enabled: true,
			ExcludePaths: []string{
				"/ping",
				"/health",
				"/favicon.ico",
			},
		},
		MediaLog: MediaConfig{
			Enabled: true,
		},
		ErrorLog: ErrorConfig{
			Enabled: true,
		},
		SystemLog: SystemConfig{
			Enabled: true,
		},
	}
}

// initializeLoggers 初始化各种类型的日志记录器
func (m *LoggerManager) initializeLoggers() {
	if !m.config.Enabled {
		return
	}
	
	// 创建基础组件
	formatter := m.createFormatter()
	writer := m.createWriter()
	
	// 初始化访问日志记录器
	if m.config.AccessLog.Enabled {
		accessLogger := NewStandardLogger()
		accessLogger.SetFormatter(formatter)
		accessLogger.SetWriter(writer)
		accessLogger.SetFilter(m.createAccessFilter())
		m.loggers[TypeAccess] = accessLogger
	}
	
	// 初始化媒体日志记录器
	if m.config.MediaLog.Enabled {
		mediaLogger := NewStandardLogger()
		mediaLogger.SetFormatter(formatter)
		mediaLogger.SetWriter(m.createMediaWriter())
		mediaLogger.SetFilter(NewMediaFilter())
		m.loggers[TypeMedia] = mediaLogger
	}
	
	// 初始化错误日志记录器
	if m.config.ErrorLog.Enabled {
		errorLogger := NewStandardLogger()
		errorLogger.SetFormatter(formatter)
		errorLogger.SetWriter(writer)
		m.loggers[TypeError] = errorLogger
	}
	
	// 初始化系统日志记录器
	if m.config.SystemLog.Enabled {
		systemLogger := NewStandardLogger()
		systemLogger.SetFormatter(formatter)
		systemLogger.SetWriter(writer)
		m.loggers[TypeSystem] = systemLogger
	}
}

// createFormatter 创建格式化器
func (m *LoggerManager) createFormatter() LogFormatter {
	switch m.config.Format {
	case "json":
		return NewJSONFormatter()
	default:
		formatter := NewTextFormatter()
		formatter.ShowColors = m.config.Output.Console.Colors
		return formatter
	}
}

// createWriter 创建通用写入器
func (m *LoggerManager) createWriter() LogWriter {
	writers := make([]LogWriter, 0)
	
	// 控制台输出
	if m.config.Output.Console.Enabled {
		consoleWriter := NewConsoleWriter()
		consoleWriter.SetFormatter(m.createFormatter())
		writers = append(writers, consoleWriter)
	}
	
	// 文件输出
	if m.config.Output.File.Enabled {
		fileWriter, err := NewFileWriter(
			m.config.Output.File.Filename,
			m.config.Output.File.MaxSize*1024*1024, // 转换为字节
			m.config.Output.File.MaxFiles,
		)
		if err != nil {
			fmt.Printf("Failed to create file writer: %v\n", err)
		} else {
			if m.config.Output.File.Async {
				asyncWriter := NewAsyncWriter(fileWriter, m.config.Output.File.BufferSize)
				writers = append(writers, asyncWriter)
			} else {
				writers = append(writers, fileWriter)
			}
		}
	}
	
	if len(writers) == 1 {
		return writers[0]
	} else if len(writers) > 1 {
		return NewMultiWriter(writers...)
	}
	
	return NewConsoleWriter() // 默认返回控制台写入器
}

// createMediaWriter 创建媒体日志专用写入器
func (m *LoggerManager) createMediaWriter() LogWriter {
	// 媒体日志可以使用单独的文件
	mediaConfig := m.config.Output.File
	if mediaConfig.Enabled {
		filename := "./logs/media_access.log"
		fileWriter, err := NewFileWriter(
			filename,
			mediaConfig.MaxSize*1024*1024,
			mediaConfig.MaxFiles,
		)
		if err != nil {
			fmt.Printf("Failed to create media file writer: %v\n", err)
			return m.createWriter() // 回退到通用写入器
		}
		
		writers := []LogWriter{fileWriter}
		
		// 如果控制台输出启用，也添加控制台输出
		if m.config.Output.Console.Enabled {
			writers = append(writers, NewConsoleWriter())
		}
		
		if len(writers) > 1 {
			return NewMultiWriter(writers...)
		}
		return fileWriter
	}
	
	return m.createWriter()
}

// createAccessFilter 创建访问日志过滤器
func (m *LoggerManager) createAccessFilter() LogFilter {
	composite := NewCompositeFilter(FilterModeAnd)
	
	// 路径过滤器
	pathFilter := NewPathFilter()
	for _, path := range m.config.AccessLog.ExcludePaths {
		pathFilter.AddExcludePath(path)
	}
	for _, path := range m.config.AccessLog.IncludePaths {
		pathFilter.AddIncludePath(path)
	}
	composite.AddFilter(pathFilter)
	
	// 状态码过滤器
	statusFilter := NewStatusCodeFilter()
	for _, code := range m.config.AccessLog.ExcludeStatus {
		statusFilter.AddExcludeStatusCode(code)
	}
	for _, code := range m.config.AccessLog.IncludeStatus {
		statusFilter.AddIncludeStatusCode(code)
	}
	composite.AddFilter(statusFilter)
	
	return composite
}

// GetLogger 获取指定类型的日志记录器
func (m *LoggerManager) GetLogger(logType LogType) Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if logger, exists := m.loggers[logType]; exists {
		return logger
	}
	
	return nil
}

// GetAccessLogger 获取访问日志记录器
func (m *LoggerManager) GetAccessLogger() Logger {
	return m.GetLogger(TypeAccess)
}

// GetMediaLogger 获取媒体日志记录器
func (m *LoggerManager) GetMediaLogger() Logger {
	return m.GetLogger(TypeMedia)
}

// GetErrorLogger 获取错误日志记录器
func (m *LoggerManager) GetErrorLogger() Logger {
	return m.GetLogger(TypeError)
}

// GetSystemLogger 获取系统日志记录器
func (m *LoggerManager) GetSystemLogger() Logger {
	return m.GetLogger(TypeSystem)
}

// UpdateConfig 更新配置并重新初始化日志记录器
func (m *LoggerManager) UpdateConfig(config *LogConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 关闭现有的日志记录器
	for _, logger := range m.loggers {
		if stdLogger, ok := logger.(*StandardLogger); ok {
			stdLogger.Close()
		}
	}
	
	// 清空现有日志记录器
	m.loggers = make(map[LogType]Logger)
	m.config = config
	
	// 重新初始化
	m.initializeLoggers()
}

// Close 关闭所有日志记录器
func (m *LoggerManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var lastErr error
	for _, logger := range m.loggers {
		if stdLogger, ok := logger.(*StandardLogger); ok {
			if err := stdLogger.Close(); err != nil {
				lastErr = err
			}
		}
	}
	
	return lastErr
}

// GetConfig 获取当前配置
func (m *LoggerManager) GetConfig() *LogConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}
