package logger

import (
	"strings"
	"github.com/gin-gonic/gin"
)

// MediaFilter 媒体文件过滤器 - 只记录图片和视频格式的访问
type MediaFilter struct {
	imageFormats []string
	videoFormats []string
}

func NewMediaFilter() *MediaFilter {
	return &MediaFilter{
		imageFormats: []string{
			".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico",
			".tiff", ".tif", ".psd", ".raw", ".cr2", ".nef", ".orf", ".sr2",
			".heic", ".heif", ".avif", ".jxl",
		},
		videoFormats: []string{
			".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm",
			".m4v", ".3gp", ".f4v", ".asf", ".rm", ".rmvb", ".vob",
			".ts", ".mts", ".m2ts", ".divx", ".xvid", ".ogv",
		},
	}
}

func (f *MediaFilter) ShouldLog(c *gin.Context, logType LogType) bool {
	// 只处理媒体类型的日志
	if logType != TypeMedia {
		return true
	}
	
	path := c.Request.URL.Path
	pathLower := strings.ToLower(path)
	
	// 检查是否为下载路径（/d/, /p/, /ad/），这些路径始终记录
	if strings.HasPrefix(path, "/d/") || strings.HasPrefix(path, "/p/") || strings.HasPrefix(path, "/ad/") {
		return true
	}
	
	// 检查是否为图片格式
	for _, format := range f.imageFormats {
		if strings.HasSuffix(pathLower, format) ||
			strings.Contains(pathLower, format+"?") ||
			strings.Contains(pathLower, format+"&") {
			return true
		}
	}
	
	// 检查是否为视频格式
	for _, format := range f.videoFormats {
		if strings.HasSuffix(pathLower, format) ||
			strings.Contains(pathLower, format+"?") ||
			strings.Contains(pathLower, format+"&") {
			return true
		}
	}
	
	return false
}

// PathFilter 路径过滤器 - 根据路径规则过滤
type PathFilter struct {
	excludePaths []string
	includePaths []string
}

func NewPathFilter() *PathFilter {
	return &PathFilter{
		excludePaths: []string{
			"/ping",
			"/health",
			"/favicon.ico",
		},
		includePaths: []string{},
	}
}

func (f *PathFilter) AddExcludePath(path string) {
	f.excludePaths = append(f.excludePaths, path)
}

func (f *PathFilter) AddIncludePath(path string) {
	f.includePaths = append(f.includePaths, path)
}

func (f *PathFilter) ShouldLog(c *gin.Context, logType LogType) bool {
	path := c.Request.URL.Path
	
	// 检查排除路径
	for _, excludePath := range f.excludePaths {
		if strings.HasPrefix(path, excludePath) {
			return false
		}
	}
	
	// 如果设置了包含路径，只记录匹配的路径
	if len(f.includePaths) > 0 {
		for _, includePath := range f.includePaths {
			if strings.HasPrefix(path, includePath) {
				return true
			}
		}
		return false
	}
	
	return true
}

// StatusCodeFilter 状态码过滤器
type StatusCodeFilter struct {
	excludeStatusCodes []int
	includeStatusCodes []int
}

func NewStatusCodeFilter() *StatusCodeFilter {
	return &StatusCodeFilter{
		excludeStatusCodes: []int{},
		includeStatusCodes: []int{},
	}
}

func (f *StatusCodeFilter) AddExcludeStatusCode(code int) {
	f.excludeStatusCodes = append(f.excludeStatusCodes, code)
}

func (f *StatusCodeFilter) AddIncludeStatusCode(code int) {
	f.includeStatusCodes = append(f.includeStatusCodes, code)
}

func (f *StatusCodeFilter) ShouldLog(c *gin.Context, logType LogType) bool {
	// 状态码过滤器主要用于访问日志
	if logType != TypeAccess {
		return true
	}
	
	statusCode := c.Writer.Status()
	
	// 检查排除状态码
	for _, excludeCode := range f.excludeStatusCodes {
		if statusCode == excludeCode {
			return false
		}
	}
	
	// 如果设置了包含状态码，只记录匹配的状态码
	if len(f.includeStatusCodes) > 0 {
		for _, includeCode := range f.includeStatusCodes {
			if statusCode == includeCode {
				return true
			}
		}
		return false
	}
	
	return true
}

// CompositeFilter 组合过滤器 - 支持多个过滤器的组合
type CompositeFilter struct {
	filters []LogFilter
	mode    FilterMode
}

type FilterMode int

const (
	FilterModeAnd FilterMode = iota // 所有过滤器都通过才记录
	FilterModeOr                    // 任一过滤器通过就记录
)

func NewCompositeFilter(mode FilterMode) *CompositeFilter {
	return &CompositeFilter{
		filters: make([]LogFilter, 0),
		mode:    mode,
	}
}

func (f *CompositeFilter) AddFilter(filter LogFilter) {
	f.filters = append(f.filters, filter)
}

func (f *CompositeFilter) ShouldLog(c *gin.Context, logType LogType) bool {
	if len(f.filters) == 0 {
		return true
	}
	
	switch f.mode {
	case FilterModeAnd:
		// 所有过滤器都必须通过
		for _, filter := range f.filters {
			if !filter.ShouldLog(c, logType) {
				return false
			}
		}
		return true
		
	case FilterModeOr:
		// 任一过滤器通过即可
		for _, filter := range f.filters {
			if filter.ShouldLog(c, logType) {
				return true
			}
		}
		return false
		
	default:
		return true
	}
}
