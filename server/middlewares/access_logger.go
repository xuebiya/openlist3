package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 常用图片格式列表
var imageFormats = []string{
	".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", 
	".webp", ".svg", ".ico", ".heic", ".avif", ".raw",
}

// 常用视频格式列表
var videoFormats = []string{
	".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", 
	".m4v", ".3gp", ".f4v", ".asf", ".rm", ".rmvb", ".vob",
	".ts", ".mts", ".m2ts", ".divx", ".xvid", ".ogv",
}

// shouldLogAccess 判断是否应该记录访问日志（只记录图片和视频格式）
func shouldLogAccess(path string) bool {
	pathLower := strings.ToLower(path)
	
	// 检查是否为下载路径（/d/, /p/, /ad/），这些路径始终记录
	if strings.HasPrefix(path, "/d/") || strings.HasPrefix(path, "/p/") || strings.HasPrefix(path, "/ad/") {
		return true
	}
	
	// 检查路径中是否包含媒体文件扩展名
	// 检查是否为图片格式
	for _, format := range imageFormats {
		if strings.HasSuffix(pathLower, format) || 
		   strings.Contains(pathLower, format+"?") || 
		   strings.Contains(pathLower, format+"&") {
			return true
		}
	}
	
	// 检查是否为视频格式
	for _, format := range videoFormats {
		if strings.HasSuffix(pathLower, format) || 
		   strings.Contains(pathLower, format+"?") || 
		   strings.Contains(pathLower, format+"&") {
			return true
		}
	}

	return false
}

// AccessLogger 用户访问日志中间件
// 实时输出用户访问日志，仅记录图片和视频格式的访问
// 格式：时间:2025年X月X日 IP地址:X.X.X.X 用户:XXX 访问路径:XXX
func AccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()
		
		// 获取访问路径（在处理请求前）
		path := c.Request.URL.Path
		
		// 预先检查是否需要记录日志，避免不必要的处理
		if !shouldLogAccess(path) {
			c.Next()
			return
		}
		
		// 处理请求
		c.Next()
		
		// 获取响应状态码
		statusCode := c.Writer.Status()
		
		// 获取当前时间并格式化为中文日期格式
		now := time.Now()
		timeStr := fmt.Sprintf("%d年%d月%d日 %02d:%02d:%02d", 
			now.Year(), int(now.Month()), now.Day(),
			now.Hour(), now.Minute(), now.Second())
		
		// 获取客户端IP地址
		clientIP := c.ClientIP()
		
		// 获取用户信息
		var username string
		userValue := c.Request.Context().Value(conf.UserKey)
		if userValue != nil {
			user, ok := userValue.(*model.User)
			if ok {
				if user.IsGuest() {
					username = "访客"
				} else {
					username = user.Username
				}
			} else {
				username = "未知用户"
			}
		} else {
			username = "未认证"
		}
		
		// 获取完整的访问路径（包含查询参数）
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}
		
		// 获取请求方法
		method := c.Request.Method
		
		// 计算请求处理时间
		duration := time.Since(startTime)
		
		// 输出访问日志到标准输出（实时显示）
		accessLog := fmt.Sprintf("时间:%s IP地址:%s 用户:%s 访问路径:%s 方法:%s 状态:%d 耗时:%v",
			timeStr, clientIP, username, path, method, statusCode, duration)
		
		// 同时输出到控制台和日志文件
		fmt.Println(accessLog)
		log.Info(accessLog)
	}
}
