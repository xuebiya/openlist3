package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/OpenListTeam/OpenList/v4/server/middlewares/logger"
)


// AccessLogger 用户访问日志中间件 - 使用新的统一日志系统
// 实时输出用户访问日志，仅记录图片和视频格式的访问
// 格式：时间:2025年X月X日 IP地址:X.X.X.X 用户:XXX 访问路径:XXX
func AccessLogger() gin.HandlerFunc {
	return MediaLogger()
}

// UserContextMiddleware 用户上下文中间件 - 设置用户信息到Gin context
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.SetupUserContext(c)
		c.Next()
	}
}

// DeprecatedAccessLogger 保留原有的访问日志中间件实现，但标记为废弃
// 这个函数保留是为了向后兼容，建议使用 UnifiedLogger() 代替
func DeprecatedAccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置用户上下文
		logger.SetupUserContext(c)
		
		// 使用媒体日志中间件
		MediaLogger()(c)
	}
}

// SetupLogging 设置日志系统 - 在应用启动时调用
func SetupLogging() {
	// 使用集成函数创建与OpenList配置兼容的日志配置
	config := logger.CreateConfigFromOpenListConfig()
	InitializeLogger(config)
	
	// 记录系统启动日志
	LogInfo("新日志系统初始化完成", map[string]interface{}{
		"version": "2.0",
		"features": []string{"访问日志", "媒体日志", "错误日志", "系统日志"},
		"config_source": "OpenList配置文件",
	})
}