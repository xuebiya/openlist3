package middlewares

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/OpenListTeam/OpenList/v4/server/middlewares/logger"
)


// 全局日志管理器
var globalLoggerManager *logger.LoggerManager

// InitializeLogger 初始化全局日志管理器
func InitializeLogger(config *logger.LogConfig) {
	if config == nil {
		config = logger.DefaultLogConfig()
	}
	globalLoggerManager = logger.NewLoggerManager(config)
}

// GetLoggerManager 获取全局日志管理器
func GetLoggerManager() *logger.LoggerManager {
	if globalLoggerManager == nil {
		InitializeLogger(nil)
	}
	return globalLoggerManager
}

// UnifiedLogger 统一日志中间件
func UnifiedLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 计算响应时间
		responseTime := time.Since(start)
		statusCode := c.Writer.Status()
		
		// 获取日志管理器
		manager := GetLoggerManager()
		if manager == nil {
			return
		}
		
		// 记录访问日志
		if accessLogger := manager.GetAccessLogger(); accessLogger != nil {
			accessLogger.LogAccess(c, statusCode, responseTime)
		}
		
		// 如果有错误，记录错误日志
		if len(c.Errors) > 0 {
			if errorLogger := manager.GetErrorLogger(); errorLogger != nil {
				for _, err := range c.Errors {
					errorLogger.LogError(c, err.Err, err.Meta.(string))
				}
			}
		}
	})
}

// MediaLogger 媒体文件访问日志中间件
func MediaLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取日志管理器
		manager := GetLoggerManager()
		if manager == nil {
			c.Next()
			return
		}
		
		mediaLogger := manager.GetMediaLogger()
		if mediaLogger == nil {
			c.Next()
			return
		}
		
		// 处理请求
		c.Next()
		
		// 获取用户名
		username := ""
		if user, exists := c.Get("username"); exists {
			if str, ok := user.(string); ok {
				username = str
			}
		}
		
		// 获取文件大小（如果可用）
		var fileSize int64
		if size, exists := c.Get("file_size"); exists {
			if s, ok := size.(int64); ok {
				fileSize = s
			}
		}
		
		// 记录媒体访问日志
		mediaLogger.LogMedia(c, username, fileSize)
	})
}

// ErrorLogger 错误日志记录函数
func LogError(c *gin.Context, err error, message string) {
	manager := GetLoggerManager()
	if manager == nil {
		return
	}
	
	if errorLogger := manager.GetErrorLogger(); errorLogger != nil {
		errorLogger.LogError(c, err, message)
	}
}

// SystemLogger 系统日志记录函数
func LogSystem(level logger.LogLevel, message string, extra map[string]interface{}) {
	manager := GetLoggerManager()
	if manager == nil {
		return
	}
	
	if systemLogger := manager.GetSystemLogger(); systemLogger != nil {
		systemLogger.LogSystem(level, message, extra)
	}
}

// LogInfo 记录信息级别的系统日志
func LogInfo(message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	LogSystem(logger.LevelInfo, message, extraData)
}

// LogWarn 记录警告级别的系统日志
func LogWarn(message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	LogSystem(logger.LevelWarn, message, extraData)
}

// LogDebug 记录调试级别的系统日志
func LogDebug(message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	LogSystem(logger.LevelDebug, message, extraData)
}

// LogSystemError 记录系统错误日志
func LogSystemError(message string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	LogSystem(logger.LevelError, message, extraData)
}

// UpdateLoggerConfig 更新日志配置
func UpdateLoggerConfig(config *logger.LogConfig) {
	if globalLoggerManager != nil {
		globalLoggerManager.UpdateConfig(config)
	} else {
		InitializeLogger(config)
	}
}

// CloseLogger 关闭全局日志管理器
func CloseLogger() {
	if globalLoggerManager != nil {
		globalLoggerManager.Close()
		globalLoggerManager = nil
	}
}
