package logger

import (
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/gin-gonic/gin"
)

// SetupUserContext 设置用户上下文信息到Gin context
func SetupUserContext(c *gin.Context) {
	// 从原有的系统获取用户信息
	userValue := c.Request.Context().Value(conf.UserKey)
	if userValue != nil {
		user, ok := userValue.(*model.User)
		if ok {
			if user.IsGuest() {
				c.Set("username", "访客")
			} else {
				c.Set("username", user.Username)
			}
		} else {
			c.Set("username", "未知用户")
		}
	} else {
		c.Set("username", "未认证")
	}
}

// CreateConfigFromOpenListConfig 从OpenList配置创建日志配置
func CreateConfigFromOpenListConfig() *LogConfig {
	config := DefaultLogConfig()
	
	// 根据OpenList的配置调整日志配置
	if conf.Conf.Log.Enable {
		config.Enabled = true
		
		// 文件配置
		config.Output.File.Enabled = true
		config.Output.File.Filename = conf.Conf.Log.Name
		config.Output.File.MaxSize = int64(conf.Conf.Log.MaxSize) // MB
		config.Output.File.MaxFiles = conf.Conf.Log.MaxBackups
		
		// 基于过滤器配置调整访问日志
		if conf.Conf.Log.Filter.Enable {
			for _, filter := range conf.Conf.Log.Filter.Filters {
				if filter.Path != "" {
					config.AccessLog.ExcludePaths = append(config.AccessLog.ExcludePaths, filter.Path)
				}
			}
		}
	} else {
		config.Enabled = false
	}
	
	return config
}

// CreateEnhancedMediaFilter 创建增强的媒体过滤器，包含原有逻辑
func CreateEnhancedMediaFilter() LogFilter {
	mediaFilter := NewMediaFilter()
	
	// 可以在这里添加额外的过滤逻辑
	// 例如，添加更多的文件格式或特殊路径
	
	return mediaFilter
}

// IntegrateWithOpenList 与OpenList系统集成
func IntegrateWithOpenList() *LoggerManager {
	config := CreateConfigFromOpenListConfig()
	manager := NewLoggerManager(config)
	
	// 设置特殊的媒体过滤器
	if mediaLogger, ok := manager.GetMediaLogger().(*StandardLogger); ok {
		mediaLogger.SetFilter(CreateEnhancedMediaFilter())
	}
	
	return manager
}
